package gitstore

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

type RemoteAccess struct {
	URL  string
	Auth transport.AuthMethod
}

type urlCredentials struct {
	Username    string
	Password    string
	HasUserinfo bool
}

type remoteInfo struct {
	Scheme   string
	Host     string
	User     string
	IsRemote bool
}

var scpLikePattern = regexp.MustCompile(`^[^@]+@[^:]+:`)

func ResolveRemoteAccess(origin string) (RemoteAccess, error) {
	cleaned := strings.TrimSpace(origin)
	if cleaned == "" {
		return RemoteAccess{}, fmt.Errorf("origin is required")
	}

	stripped, creds, err := stripCredentials(cleaned)
	if err != nil {
		return RemoteAccess{}, err
	}

	if !isRemoteOrigin(stripped) {
		return RemoteAccess{URL: stripped}, nil
	}

	rewritten, _, err := applyInsteadOf(stripped)
	if err != nil {
		return RemoteAccess{}, err
	}

	auth, err := resolveAuth(rewritten, creds)
	if err != nil {
		return RemoteAccess{}, err
	}

	return RemoteAccess{URL: rewritten, Auth: auth}, nil
}

func resolveAuth(origin string, creds urlCredentials) (transport.AuthMethod, error) {
	info, err := parseRemoteInfo(origin)
	if err != nil {
		return nil, err
	}
	if !info.IsRemote {
		return nil, nil
	}

	switch info.Scheme {
	case "http", "https":
		return resolveHTTPAuth(info.Host, creds)
	case "ssh":
		return resolveSSHAuth(info.User)
	default:
		return nil, nil
	}
}

func resolveHTTPAuth(host string, creds urlCredentials) (transport.AuthMethod, error) {
	if creds.HasUserinfo {
		return &githttp.BasicAuth{Username: creds.Username, Password: creds.Password}, nil
	}
	if username, password, ok := githubTokenAuth(host); ok {
		return &githttp.BasicAuth{Username: username, Password: password}, nil
	}
	if username, password, ok, err := netrcCredentials(host); err != nil {
		return nil, err
	} else if ok {
		return &githttp.BasicAuth{Username: username, Password: password}, nil
	}
	if username, password, ok := envBasicAuth(); ok {
		return &githttp.BasicAuth{Username: username, Password: password}, nil
	}
	return nil, nil
}

func resolveSSHAuth(user string) (transport.AuthMethod, error) {
	auth, err := gitssh.NewSSHAgentAuth(user)
	if err != nil {
		return nil, fmt.Errorf("ssh agent auth: %w", err)
	}
	if isEnvTrue("ASM_SSH_INSECURE") {
		auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	return auth, nil
}

func githubTokenAuth(host string) (string, string, bool) {
	if !strings.EqualFold(host, "github.com") {
		return "", "", false
	}
	for _, key := range []string{"ASM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		if token := os.Getenv(key); token != "" {
			return "x-access-token", token, true
		}
	}
	return "", "", false
}

func envBasicAuth() (string, string, bool) {
	if token := os.Getenv("ASM_GIT_TOKEN"); token != "" {
		username := os.Getenv("ASM_GIT_USERNAME")
		if username == "" {
			username = "x-access-token"
		}
		return username, token, true
	}

	username := os.Getenv("ASM_GIT_USERNAME")
	password := os.Getenv("ASM_GIT_PASSWORD")
	if username != "" || password != "" {
		return username, password, true
	}
	return "", "", false
}

func parseRemoteInfo(origin string) (remoteInfo, error) {
	if _, ok := schemeForOrigin(origin); ok {
		parsed, err := url.Parse(origin)
		if err != nil {
			return remoteInfo{}, err
		}
		info := remoteInfo{
			Scheme:   strings.ToLower(parsed.Scheme),
			Host:     strings.ToLower(parsed.Hostname()),
			IsRemote: true,
		}
		if parsed.User != nil {
			info.User = parsed.User.Username()
		}
		return info, nil
	}

	if scpLikePattern.MatchString(origin) {
		parts := strings.SplitN(origin, "@", 2)
		hostPart := parts[1]
		host := strings.SplitN(hostPart, ":", 2)[0]
		return remoteInfo{
			Scheme:   "ssh",
			Host:     strings.ToLower(host),
			User:     parts[0],
			IsRemote: true,
		}, nil
	}

	return remoteInfo{IsRemote: false}, nil
}

func stripCredentials(origin string) (string, urlCredentials, error) {
	creds := urlCredentials{}
	scheme, ok := schemeForOrigin(origin)
	if !ok {
		return origin, creds, nil
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return "", creds, err
	}

	if parsed.User != nil {
		creds.Username = parsed.User.Username()
		if password, ok := parsed.User.Password(); ok {
			creds.Password = password
		}
		if creds.Username != "" || creds.Password != "" {
			creds.HasUserinfo = true
		}
	}

	switch scheme {
	case "ssh":
		if parsed.User != nil {
			if creds.Username != "" {
				parsed.User = url.User(creds.Username)
			} else {
				parsed.User = nil
			}
		}
	default:
		parsed.User = nil
	}

	return parsed.String(), creds, nil
}

func schemeForOrigin(origin string) (string, bool) {
	index := strings.Index(origin, "://")
	if index <= 0 {
		return "", false
	}
	return strings.ToLower(origin[:index]), true
}

func isRemoteOrigin(origin string) bool {
	if scheme, ok := schemeForOrigin(origin); ok {
		switch scheme {
		case "git", "http", "https", "ssh":
			return true
		default:
			return false
		}
	}
	return scpLikePattern.MatchString(origin)
}

func isEnvTrue(name string) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
