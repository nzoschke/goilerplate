package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type generator struct {
	name   string
	bin    string
	args   []string
	skipFn func() bool
	runFn  func() error // custom run function (if set, bin/args ignored)
}

func GenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gen",
		Short: "Run code generators (templ, tailwind) in parallel",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGen()
		},
	}
}

func runGen() error {
	required := []string{"tailwindcss", "templ"}
	var missing []string
	for _, bin := range required {
		if _, err := exec.LookPath(bin); err != nil {
			missing = append(missing, bin)
		}
	}
	if len(missing) > 0 {
		fmt.Println("Missing binaries:", missing)
		fmt.Println("Install with:")
		fmt.Println("  go install github.com/a-h/templ/cmd/templ@latest")
		fmt.Println("  # tailwindcss: https://tailwindcss.com/blog/standalone-cli")
		return fmt.Errorf("missing required binaries: %v", missing)
	}

	generators := []generator{
		{
			name:   "tailwindcss",
			bin:    "tailwindcss",
			args:   []string{"-i", "assets/css/input.css", "-o", "assets/css/output.css", "--minify"},
			skipFn: skipTailwind,
		},
		{
			name:   "templ",
			bin:    "templ",
			args:   []string{"generate"},
			skipFn: skipTempl,
		},
		{
			name:   "jukebox",
			skipFn: skipJukebox,
			runFn:  runJukebox,
		},
	}

	start := time.Now()
	var wg sync.WaitGroup
	errCh := make(chan error, len(generators))

	for _, g := range generators {
		wg.Add(1)
		go func(g generator) {
			defer wg.Done()

			if g.skipFn != nil && g.skipFn() {
				fmt.Printf("[%s] skipped\n", g.name)
				return
			}

			genStart := time.Now()
			var err error
			if g.runFn != nil {
				err = g.runFn()
			} else {
				cmd := exec.Command(g.bin, g.args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
			}

			if err != nil {
				errCh <- fmt.Errorf("%s: %w", g.name, err)
				return
			}

			fmt.Printf("[%s] done (%s)\n", g.name, time.Since(genStart).Round(time.Millisecond))
		}(g)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println("error:", err)
		}
		return fmt.Errorf("generation failed")
	}

	fmt.Printf("done (%s)\n", time.Since(start).Round(time.Millisecond))
	return nil
}

func skipTailwind() bool {
	inputs := []string{"assets/css/input.css"}
	_ = filepath.WalkDir("internal/ui", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".templ") || strings.HasSuffix(path, ".go") {
			inputs = append(inputs, path)
		}
		return nil
	})
	jsFiles, _ := filepath.Glob("assets/js/*.js")
	inputs = append(inputs, jsFiles...)
	return isUpToDate("assets/css/output.css", inputs)
}

func skipTempl() bool {
	var templFiles []string
	_ = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".templ") {
			templFiles = append(templFiles, path)
		}
		return nil
	})

	for _, templFile := range templFiles {
		outFile := strings.TrimSuffix(templFile, ".templ") + "_templ.go"
		if !isUpToDate(outFile, []string{templFile}) {
			return false
		}
	}
	return true
}

func isUpToDate(output string, inputs []string) bool {
	outInfo, err := os.Stat(output)
	if err != nil {
		return false
	}
	outMod := outInfo.ModTime()

	for _, input := range inputs {
		inInfo, err := os.Stat(input)
		if err != nil {
			continue
		}
		if inInfo.ModTime().After(outMod) {
			return false
		}
	}
	return true
}

func skipJukebox() bool {
	// Check if submodule exists
	if _, err := os.Stat("jukelab"); os.IsNotExist(err) {
		return true // skip if no submodule
	}

	// Get current submodule commit
	cmd := exec.Command("git", "-C", "jukelab", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return false // rebuild if can't get commit
	}
	currentCommit := strings.TrimSpace(string(out))

	// Check stored commit
	commitFile := "jukelab/build/.commit"
	stored, err := os.ReadFile(commitFile)
	if err != nil {
		return false // rebuild if no stored commit
	}

	return strings.TrimSpace(string(stored)) == currentCommit
}

func runJukebox() error {
	jukelabDir := "jukelab"

	// Check for npm
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("missing required binary: npm")
	}

	// Install dependencies
	installCmd := exec.Command("npm", "install")
	installCmd.Dir = jukelabDir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	// Build with BASE_PATH
	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = jukelabDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Env = append(os.Environ(), "BASE_PATH=/jukebox")
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("npm run build failed: %w", err)
	}

	// Store current commit
	cmd := exec.Command("git", "-C", jukelabDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}
	if err := os.WriteFile("jukelab/build/.commit", out, 0644); err != nil {
		return fmt.Errorf("failed to write commit file: %w", err)
	}

	return nil
}
