package asm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

const (
	defaultSkillsAPIBase = "https://skills.sh"
	defaultFindLimit     = 10
)

type searchSkill struct {
	ID        string `json:"id"`
	SkillID   string `json:"skillId"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	TopSource string `json:"topSource"`
	Installs  int    `json:"installs"`
}

type searchResponse struct {
	Skills []searchSkill `json:"skills"`
}

func Find(query string) (FindReport, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return FindReport{}, fmt.Errorf("query is required")
	}

	skills, err := searchSkillsAPI(query, defaultFindLimit)
	if err != nil {
		return FindReport{}, err
	}

	return FindReport{Query: query, Skills: skills}, nil
}

func searchSkillsAPI(query string, limit int) ([]FindSkill, error) {
	base := skillsAPIBase()
	endpoint := strings.TrimRight(base, "/") + "/api/search"
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid skills api url %q: %w", base, err)
	}

	params := parsed.Query()
	params.Set("q", query)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	parsed.RawQuery = params.Encode()

	debug.Logf("find query=%q url=%s", query, parsed.String())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(parsed.String())
	if err != nil {
		return nil, fmt.Errorf("search skills: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search skills: unexpected status %s", resp.Status)
	}

	var payload searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("search skills: decode response: %w", err)
	}

	results := make([]FindSkill, 0, len(payload.Skills))
	for _, skill := range payload.Skills {
		normalized, ok := normalizeSearchSkill(skill)
		if !ok {
			continue
		}
		results = append(results, normalized)
	}

	debug.Logf("find results=%d", len(results))
	return results, nil
}

func normalizeSearchSkill(skill searchSkill) (FindSkill, bool) {
	source := strings.TrimSpace(skill.Source)
	if source == "" {
		source = strings.TrimSpace(skill.TopSource)
	}

	skillID := strings.TrimSpace(skill.SkillID)
	if skillID == "" {
		skillID = strings.TrimSpace(skill.Name)
	}

	if (source == "" || skillID == "") && skill.ID != "" {
		parts := strings.Split(skill.ID, "/")
		if len(parts) >= 3 {
			if source == "" {
				source = strings.Join(parts[:2], "/")
			}
			if skillID == "" {
				skillID = parts[len(parts)-1]
			}
		}
	}

	if source == "" || skillID == "" {
		return FindSkill{}, false
	}

	return FindSkill{
		Source:   source,
		SkillID:  skillID,
		Name:     strings.TrimSpace(skill.Name),
		Installs: skill.Installs,
	}, true
}

func skillsAPIBase() string {
	if value := strings.TrimSpace(os.Getenv("SKILLS_API_URL")); value != "" {
		return value
	}
	return defaultSkillsAPIBase
}
