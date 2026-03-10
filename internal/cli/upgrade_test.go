package cli

import "testing"

func TestUpgradeCommandRegistered(t *testing.T) {
	cmd, _, _ := newTestCommand()

	found, _, err := cmd.Find([]string{"upgrade"})
	if err != nil {
		t.Fatalf("find upgrade: %v", err)
	}
	if found == nil {
		t.Fatalf("expected upgrade command")
	}
	if found.Name() != "upgrade" {
		t.Fatalf("expected command name %q, got %q", "upgrade", found.Name())
	}
	if found.Short != "Upgrade asm via go install" {
		t.Fatalf("unexpected short help: %q", found.Short)
	}
}
