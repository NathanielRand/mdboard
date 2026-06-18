package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Fetch and install the latest version of mdboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Fetching latest mdboard...")
		c := exec.Command("go", "install", "github.com/NathanielRand/mdboard@latest")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("upgrade failed: %w", err)
		}

		// go install writes to $GOPATH/bin, which may not be in PATH.
		// Copy the freshly-built binary to the install location so the
		// running binary is actually updated.
		gopath, err := exec.Command("go", "env", "GOPATH").Output()
		if err != nil {
			return fmt.Errorf("could not determine GOPATH: %w", err)
		}
		src := filepath.Join(strings.TrimSpace(string(gopath)), "bin", "mdboard")
		cp := exec.Command("sudo", "cp", src, binaryPath)
		cp.Stdout = os.Stdout
		cp.Stderr = os.Stderr
		if err := cp.Run(); err != nil {
			return fmt.Errorf("upgraded binary built at %s but failed to copy to %s: %w", src, binaryPath, err)
		}

		fmt.Println("✅ mdboard upgraded successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
