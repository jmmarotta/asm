package cli

import "testing"

func TestCommandAliases(t *testing.T) {
	cmd, _, _ := newTestCommand()

	tests := []struct {
		alias    string
		expected string
	}{
		{alias: "a", expected: "add"},
		{alias: "i", expected: "install"},
		{alias: "rm", expected: "remove"},
		{alias: "uninstall", expected: "remove"},
		{alias: "up", expected: "update"},
	}

	for _, test := range tests {
		found, _, err := cmd.Find([]string{test.alias})
		if err != nil {
			t.Fatalf("find alias %q: %v", test.alias, err)
		}
		if found == nil {
			t.Fatalf("find alias %q: command not found", test.alias)
		}
		if found.Name() != test.expected {
			t.Fatalf("alias %q resolved to %q", test.alias, found.Name())
		}
	}
}
