package cmd

import (
	"fmt"
	"os"
	"os/exec"

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
	jukelabDir := "jukelab"

	// Check for required tools
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("missing required binary: npm")
	}

	// Check submodule exists
	if _, err := os.Stat(jukelabDir); os.IsNotExist(err) {
		return fmt.Errorf("jukelab submodule not found - run: git submodule update --init")
	}

	fmt.Println("==> Building JukeLab for /jukebox...")

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

	fmt.Println("==> Done!")
	return nil
}

func runIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
