package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/asm"
	"github.com/jmmarotta/agent_skills_manager/internal/linker"
)

func TestPrintInstallReportNoSkills(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printInstallReport(asm.InstallReport{NoSkills: true}, &out, &errOut)

	if out.String() != "No skills found.\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
	if errOut.String() != "" {
		t.Fatalf("expected no warnings, got %q", errOut.String())
	}
}

func TestPrintInstallReportWarnings(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	report := asm.InstallReport{
		Linked: 2,
		Pruned: 1,
		Warnings: []linker.Warning{
			{Message: "first"},
			{Message: "second", Target: "/tmp/skill"},
		},
	}

	printInstallReport(report, &out, &errOut)

	if !strings.Contains(out.String(), "Installed: 2, Pruned: 1, Warnings: 2") {
		t.Fatalf("unexpected summary output: %q", out.String())
	}
	if !strings.Contains(errOut.String(), "warning: first") {
		t.Fatalf("missing warning: %q", errOut.String())
	}
	if !strings.Contains(errOut.String(), "warning: second (/tmp/skill)") {
		t.Fatalf("missing targeted warning: %q", errOut.String())
	}
}

func TestPrintUpdateReportNoSkills(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printUpdateReport(asm.UpdateReport{
		Install:        asm.InstallReport{NoSkills: true},
		UpdatedOrigins: []string{"origin-a"},
	}, &out, &errOut)

	if out.String() != "No skills found.\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
	if strings.Contains(out.String(), "Updated origins:") {
		t.Fatalf("did not expect updated origins")
	}
}

func TestPrintUpdateReportOrigins(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printUpdateReport(asm.UpdateReport{
		Install: asm.InstallReport{Linked: 1, Pruned: 0},
		UpdatedOrigins: []string{
			"origin-a",
			"origin-b",
		},
	}, &out, &errOut)

	if !strings.Contains(out.String(), "Installed: 1, Pruned: 0, Warnings: 0") {
		t.Fatalf("missing install summary: %q", out.String())
	}
	if !strings.Contains(out.String(), "Updated origins: origin-a, origin-b") {
		t.Fatalf("missing updated origins: %q", out.String())
	}
}

func TestPrintUpdateReportEmptyOrigins(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printUpdateReport(asm.UpdateReport{
		Install:        asm.InstallReport{Linked: 1, Pruned: 0},
		UpdatedOrigins: nil,
	}, &out, &errOut)

	if strings.Contains(out.String(), "Updated origins:") {
		t.Fatalf("did not expect updated origins: %q", out.String())
	}
}

func TestPrintRemoveReport(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printRemoveReport(asm.RemoveReport{
		Install: asm.InstallReport{Linked: 0, Pruned: 1},
		Removed: []asm.SkillSummary{{
			Name:   "foo",
			Origin: "origin-a",
		}},
		PrunedStores: []string{"origin-a"},
	}, &out, &errOut)

	if !strings.Contains(out.String(), "Installed: 0, Pruned: 1, Warnings: 0") {
		t.Fatalf("missing install summary: %q", out.String())
	}
	if !strings.Contains(out.String(), "Removed: foo (origin-a)") {
		t.Fatalf("missing removed line: %q", out.String())
	}
	if !strings.Contains(out.String(), "Pruned store: origin-a") {
		t.Fatalf("missing pruned store line: %q", out.String())
	}
}

func TestPrintRemoveReportNoOrigin(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printRemoveReport(asm.RemoveReport{
		Install: asm.InstallReport{Linked: 0, Pruned: 0},
		Removed: []asm.SkillSummary{{Name: "foo"}},
	}, &out, &errOut)

	if !strings.Contains(out.String(), "Removed: foo") {
		t.Fatalf("missing removed line: %q", out.String())
	}
	if strings.Contains(out.String(), "Pruned store:") {
		t.Fatalf("did not expect pruned store line: %q", out.String())
	}
}

func TestPrintRemoveReportNoChanges(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	printRemoveReport(asm.RemoveReport{
		Warnings:  []string{"skill \"foo\" not found"},
		NoChanges: true,
	}, &out, &errOut)

	if out.String() != "No matching skills removed.\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
	if errOut.String() != "warning: skill \"foo\" not found\n" {
		t.Fatalf("unexpected warning output: %q", errOut.String())
	}
}

func TestPrintListReportNoSkills(t *testing.T) {
	var out bytes.Buffer

	if err := printListReport(asm.ListReport{NoSkills: true}, &out); err != nil {
		t.Fatalf("print list: %v", err)
	}
	if out.String() != "No skills found.\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestPrintListReport(t *testing.T) {
	var out bytes.Buffer

	report := asm.ListReport{
		Skills: []asm.SkillSummary{{
			Name:    "foo",
			Type:    "path",
			Origin:  "/tmp/skill",
			Version: "",
			Subdir:  "",
		}},
	}

	if err := printListReport(report, &out); err != nil {
		t.Fatalf("print list: %v", err)
	}
	if !strings.Contains(out.String(), "NAME") || !strings.Contains(out.String(), "TYPE") {
		t.Fatalf("missing header: %q", out.String())
	}
	if !strings.Contains(out.String(), "foo") {
		t.Fatalf("missing skill row: %q", out.String())
	}
}

func TestPrintShowReport(t *testing.T) {
	var out bytes.Buffer

	report := asm.ShowReport{
		Name:    "foo",
		Type:    "git",
		Origin:  "origin-a",
		Subdir:  "bar",
		Version: "v1.0.0",
		Replace: "/tmp/repo",
	}

	if err := printShowReport(report, &out); err != nil {
		t.Fatalf("print show: %v", err)
	}

	var decoded asm.ShowReport
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Name != "foo" || decoded.Origin != "origin-a" {
		t.Fatalf("unexpected decoded report: %+v", decoded)
	}
}

func TestPrintInitReport(t *testing.T) {
	var out bytes.Buffer
	printInitReport(asm.InitReport{}, &out)
	if out.String() != "Initialized skills.jsonc\n" {
		t.Fatalf("unexpected init output: %q", out.String())
	}
}
