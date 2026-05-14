package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPhaseIndex(t *testing.T) {
	cases := map[string]int{
		"research": 0,
		"generate": 1,
		"build":    2,
		"verify":   3,
		"publish":  4,
		"missing":  -1,
	}
	for phase, want := range cases {
		if got := phaseIndex(phase); got != want {
			t.Fatalf("phaseIndex(%q) = %d, want %d", phase, got, want)
		}
	}
}

func TestStandalonePhaseHelpers(t *testing.T) {
	home, err := os.MkdirTemp("/home/shubhan/tmp", "pp-test-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, "tmp"), 0o755); err != nil {
		t.Fatalf("mkdir tmp: %v", err)
	}

	api := "Notion API"
	slug := "notion-api"
	runDir := filepath.Join(home, "hermes-press", "runs", slug)
	libraryDir := filepath.Join(home, "hermes-press", "library")

	research, err := execResearch(api, "spec.yaml", "https://example.com/api", runDir)
	if err != nil {
		t.Fatalf("execResearch returned error: %v", err)
	}
	if research.Slug != slug {
		t.Fatalf("execResearch slug = %q, want %q", research.Slug, slug)
	}
	if _, err := os.Stat(filepath.Join(runDir, "research.json")); err != nil {
		t.Fatalf("research.json missing: %v", err)
	}

	generate, err := execGenerate(slug, runDir, false)
	if err != nil {
		t.Fatalf("execGenerate returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(generate.WorkDir, "cmd", slug+"-cli", "main.go")); err != nil {
		t.Fatalf("generated main.go missing: %v", err)
	}

	if err := os.WriteFile(filepath.Join(generate.WorkDir, "go.mod"), []byte("module example.com/"+slug+"\n\ngo 1.20\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(generate.WorkDir, "main.go"), []byte("package main\n\nimport \"flag\"\n\nfunc main() { flag.Parse() }\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	binaryPath := filepath.Join(generate.WorkDir, slug+"-cli")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Dir = generate.WorkDir
	build.Env = append(os.Environ(), "GOTMPDIR="+filepath.Join(home, "tmp"))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build binary stub: %v\n%s", err, out)
	}

	verify, err := execVerify(slug, generate.WorkDir, binaryPath)
	if err != nil {
		t.Fatalf("execVerify returned error: %v", err)
	}
	if !verify.Passed {
		t.Fatalf("execVerify passed = false, checks = %#v", verify.Checks)
	}

	publish, err := execPublish(slug, runDir, libraryDir, false)
	if err != nil {
		t.Fatalf("execPublish returned error: %v", err)
	}
	if publish.Status != "published" {
		t.Fatalf("publish status = %q, want %q", publish.Status, "published")
	}
	if _, err := os.Stat(filepath.Join(libraryDir, slug)); err != nil {
		t.Fatalf("published directory missing: %v", err)
	}
}
