package cmd

import (
	"fmt"
	"os"
	"os/exec"

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
		fmt.Println("✅ mdboard upgraded successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
