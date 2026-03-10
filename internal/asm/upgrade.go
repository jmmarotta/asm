package asm

import (
	"fmt"
	"os/exec"
	"strings"
)

const upgradeInstallTarget = "github.com/jmmarotta/agent_skills_manager/cmd/asm@latest"

var upgradeCommandRunner = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func Upgrade() (UpgradeReport, error) {
	output, err := upgradeCommandRunner("go", "install", upgradeInstallTarget)
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return UpgradeReport{}, fmt.Errorf("go install %s failed: %s: %w", upgradeInstallTarget, message, err)
		}
		return UpgradeReport{}, fmt.Errorf("go install %s failed: %w", upgradeInstallTarget, err)
	}

	return UpgradeReport{Target: upgradeInstallTarget}, nil
}
