package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/templui/goilerplate/cmd/do/cmd"

	"github.com/spf13/cobra"
)

func main() {
	maybeRebuild()

	rootCmd := &cobra.Command{
		Use:   "do",
		Short: "Development tools for goilerplate",
	}

	rootCmd.AddCommand(cmd.DevCmd())
	rootCmd.AddCommand(cmd.GenCmd())
	rootCmd.AddCommand(cmd.BuildCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func maybeRebuild() {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	if !strings.HasSuffix(exe, "bin/do") {
		return
	}

	binInfo, err := os.Stat(exe)
	if err != nil {
		return
	}
	binMod := binInfo.ModTime()

	var needsRebuild bool
	_ = filepath.WalkDir("cmd/do", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(binMod) {
			needsRebuild = true
			return filepath.SkipAll
		}
		return nil
	})

	if !needsRebuild {
		return
	}

	fmt.Println("Rebuilding bin/do...")
	build := exec.Command("go", "build", "-o", exe, "./cmd/do")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Println("Rebuild failed:", err)
		return
	}

	if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
		fmt.Println("Re-exec failed:", err)
	}
}
