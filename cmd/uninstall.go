package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

const binDir = "/usr/local/bin"
const binaryPath = binDir + "/mdboard"
const aliasPath = binDir + "/mdb"

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove mdboard and the mdb alias from /usr/local/bin",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This will remove:")
		fmt.Printf("  %s\n", binaryPath)
		fmt.Printf("  %s\n", aliasPath)
		fmt.Print("\nUninstall mdboard? (y/N) ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}

		targets := []string{aliasPath, binaryPath}
		for _, t := range targets {
			if _, err := os.Lstat(t); os.IsNotExist(err) {
				fmt.Printf("  (skipping %s — not found)\n", t)
				continue
			}
			c := exec.Command("sudo", "rm", "-f", t)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("failed to remove %s: %w", t, err)
			}
			fmt.Printf("  removed %s\n", t)
		}

		fmt.Println("\n✅ mdboard uninstalled.")
		fmt.Println("   Project .mdboard/config.yaml files were not touched.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
