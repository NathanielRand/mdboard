package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type checkResult struct {
	label   string
	ok      bool
	detail  string
	repair  func() error
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and repair the mdboard installation",
	RunE: func(cmd *cobra.Command, args []string) error {
		checks := buildChecks()
		allOK := true
		var broken []checkResult

		for _, c := range checks {
			if c.ok {
				fmt.Printf("  ✅ %s\n", c.label)
			} else {
				fmt.Printf("  ❌ %s", c.label)
				if c.detail != "" {
					fmt.Printf(" — %s", c.detail)
				}
				fmt.Println()
				allOK = false
				if c.repair != nil {
					broken = append(broken, c)
				}
			}
		}

		if allOK {
			fmt.Println("\nAll checks passed.")
			return nil
		}

		if len(broken) == 0 {
			fmt.Println("\nSome checks failed but no automatic repairs are available. Re-run install.sh to fix.")
			return nil
		}

		fmt.Printf("\n%d repair(s) available. Apply them now? (y/N) ", len(broken))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
			fmt.Println("Skipped repairs.")
			return nil
		}

		for _, c := range broken {
			fmt.Printf("  Repairing: %s...\n", c.label)
			if err := c.repair(); err != nil {
				fmt.Fprintf(os.Stderr, "    failed: %v\n", err)
			} else {
				fmt.Printf("    ✅ fixed\n")
			}
		}
		return nil
	},
}

func buildChecks() []checkResult {
	return []checkResult{
		checkBinary(),
		checkAlias(),
		checkPathContainsBinDir(),
		checkGoToolchain(),
	}
}

func checkBinary() checkResult {
	info, err := os.Stat(binaryPath)
	if err != nil {
		return checkResult{
			label:  "mdboard binary at " + binaryPath,
			ok:     false,
			detail: "not found — re-run install.sh",
		}
	}
	if info.Mode()&0111 == 0 {
		return checkResult{
			label:  "mdboard binary at " + binaryPath,
			ok:     false,
			detail: "not executable",
			repair: func() error {
				return exec.Command("sudo", "chmod", "+x", binaryPath).Run()
			},
		}
	}
	return checkResult{label: "mdboard binary at " + binaryPath, ok: true}
}

func checkAlias() checkResult {
	target, err := os.Readlink(aliasPath)
	if err != nil {
		// not a symlink or doesn't exist
		if _, err2 := os.Stat(aliasPath); err2 == nil {
			return checkResult{label: "mdb alias at " + aliasPath, ok: true}
		}
		return checkResult{
			label:  "mdb alias at " + aliasPath,
			ok:     false,
			detail: "missing",
			repair: func() error {
				c := exec.Command("sudo", "ln", "-sf", binaryPath, aliasPath)
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				return c.Run()
			},
		}
	}
	// symlink exists — verify it resolves
	resolved := target
	if !filepath.IsAbs(target) {
		resolved = filepath.Join(filepath.Dir(aliasPath), target)
	}
	if _, err := os.Stat(resolved); err != nil {
		return checkResult{
			label:  "mdb alias at " + aliasPath,
			ok:     false,
			detail: fmt.Sprintf("broken symlink (→ %s)", target),
			repair: func() error {
				c := exec.Command("sudo", "ln", "-sf", binaryPath, aliasPath)
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				return c.Run()
			},
		}
	}
	return checkResult{label: "mdb alias at " + aliasPath, ok: true}
}

func checkPathContainsBinDir() checkResult {
	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == binDir {
			return checkResult{label: binDir + " is in PATH", ok: true}
		}
	}
	return checkResult{
		label:  binDir + " is in PATH",
		ok:     false,
		detail: fmt.Sprintf("add 'export PATH=%s:$PATH' to your shell profile", binDir),
	}
}

func checkGoToolchain() checkResult {
	_, err := exec.LookPath("go")
	if err != nil {
		return checkResult{
			label:  "Go toolchain available",
			ok:     false,
			detail: "go not found on PATH — required for mdb upgrade",
		}
	}
	return checkResult{label: "Go toolchain available", ok: true}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
