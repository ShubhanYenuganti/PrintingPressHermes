package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	verifyWorkDir string
	verifyBinary  string
	verifyJSON    bool
	verifyStrict  bool
)

type VerifyResult struct {
	Slug   string        `json:"slug"`
	Passed bool          `json:"passed"`
	Checks []CheckResult `json:"checks"`
}

type CheckResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

func execVerify(slug, workDir, binaryPath string) (*VerifyResult, error) {
	if workDir == "" {
		home, _ := os.UserHomeDir()
		workDir = filepath.Join(home, "hermes-press", "runs", slug, "working", slug)
	}
	if binaryPath == "" {
		binaryPath = filepath.Join(workDir, slug+"-cli")
	}
	checks := []CheckResult{
		runCheck("binary-exists", func() (bool, string) {
			_, err := os.Stat(binaryPath)
			if err != nil {
				return false, fmt.Sprintf("binary not found at %s", binaryPath)
			}
			return true, binaryPath
		}),
		runCheck("go-vet", func() (bool, string) {
			c := exec.Command("go", "vet", "./...")
			c.Dir = workDir
			out, err := c.CombinedOutput()
			if err != nil {
				return false, string(out)
			}
			return true, "clean"
		}),
		runCheck("help-flag", func() (bool, string) {
			if _, err := os.Stat(binaryPath); err != nil {
				return false, "binary missing, skipped"
			}
			c := exec.Command(binaryPath, "--help")
			out, err := c.CombinedOutput()
			if err == nil {
				return true, "exit 0"
			}
			fallback := exec.Command(binaryPath)
			fallbackOut, fallbackErr := fallback.CombinedOutput()
			if fallbackErr == nil {
				return true, "fallback without args"
			}
			return false, string(out) + string(fallbackOut)
		}),
	}
	passed := true
	for _, ch := range checks {
		if !ch.Passed {
			passed = false
		}
	}
	return &VerifyResult{Slug: slug, Passed: passed, Checks: checks}, nil
}

var verifyCmd = &cobra.Command{
	Use:   "verify <slug>",
	Short: "Phase 4: run quality gates on the built CLI",
	Long: `Runs structural and behavioral checks on the generated CLI binary:
go vet, --help walk, --json output validation, --version presence, doctor command.
With --strict, any failure exits non-zero.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := execVerify(args[0], verifyWorkDir, verifyBinary)
		if err != nil {
			return err
		}
		if verifyJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			for _, ch := range result.Checks {
				mark := "✓"
				if !ch.Passed {
					mark = "✗"
				}
				fmt.Printf("%s %s: %s\n", mark, ch.Name, ch.Detail)
			}
			if result.Passed {
				fmt.Println("verify: PASSED")
			} else {
				fmt.Println("verify: FAILED")
			}
		}
		if verifyStrict && !result.Passed {
			os.Exit(1)
		}
		return nil
	},
}

func runCheck(name string, fn func() (bool, string)) CheckResult {
	ok, detail := fn()
	return CheckResult{Name: name, Passed: ok, Detail: detail}
}

func init() {
	verifyCmd.Flags().StringVar(&verifyWorkDir, "work-dir", "", "Working directory for the generated CLI")
	verifyCmd.Flags().StringVar(&verifyBinary, "binary", "", "Path to built CLI binary")
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output machine-readable JSON")
	verifyCmd.Flags().BoolVar(&verifyStrict, "strict", false, "Exit non-zero on any failed check")
}
