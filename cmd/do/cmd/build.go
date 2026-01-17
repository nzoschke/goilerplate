package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func BuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build commands",
	}

	cmd.AddCommand(buildJukeboxCmd())
	return cmd
}

func buildJukeboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jukebox",
		Short: "Build JukeLab SvelteKit app and embed in jukebox/",
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildJukebox()
		},
	}
}

func buildJukebox() error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	jukelabDir := filepath.Join(projectRoot, ".jukelab")
	jukeboxDir := filepath.Join(projectRoot, "jukebox")

	// Check for required tools
	for _, bin := range []string{"git", "npm"} {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("missing required binary: %s", bin)
		}
	}

	fmt.Println("==> Building JukeLab for /jukebox...")

	// Clone or update repo
	if _, err := os.Stat(jukelabDir); os.IsNotExist(err) {
		fmt.Println("==> Cloning jukelab repo...")
		if err := run("git", "clone", "https://github.com/nzoschke/jukelab.git", jukelabDir); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		fmt.Println("==> Updating existing jukelab repo...")
		if err := runIn(jukelabDir, "git", "fetch", "origin"); err != nil {
			return fmt.Errorf("git fetch failed: %w", err)
		}
		if err := runIn(jukelabDir, "git", "reset", "--hard", "origin/main"); err != nil {
			return fmt.Errorf("git reset failed: %w", err)
		}
	}

	// Install dependencies
	fmt.Println("==> Installing dependencies...")
	if err := runIn(jukelabDir, "npm", "install"); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	// Build with BASE_PATH
	fmt.Println("==> Building SvelteKit app with BASE_PATH=/jukebox...")
	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = jukelabDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Env = append(os.Environ(), "BASE_PATH=/jukebox")
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("npm run build failed: %w", err)
	}

	// Copy build output
	fmt.Println("==> Copying build to", jukeboxDir)
	if err := os.RemoveAll(jukeboxDir); err != nil {
		return fmt.Errorf("failed to remove old jukebox dir: %w", err)
	}
	if err := copyDir(filepath.Join(jukelabDir, "build"), jukeboxDir); err != nil {
		return fmt.Errorf("failed to copy build: %w", err)
	}

	// Create embed.go
	fmt.Println("==> Creating embed.go...")
	embedGo := `package jukebox

import "embed"

// JukeboxFS contains the built SvelteKit app files.
// Run "go run ./cmd/do build jukebox" to build and populate this directory.
//
//go:embed all:*
var JukeboxFS embed.FS
`
	if err := os.WriteFile(filepath.Join(jukeboxDir, "embed.go"), []byte(embedGo), 0644); err != nil {
		return fmt.Errorf("failed to write embed.go: %w", err)
	}

	fmt.Println("==> Done! JukeLab built and copied to jukebox/")
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}
