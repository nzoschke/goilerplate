package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

func DevCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dev",
		Short: "Run air for hot-reload development",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDev()
		},
	}
}

func runDev() error {
	airPath, err := exec.LookPath("air")
	if err != nil {
		fmt.Println("Missing binary: air")
		fmt.Println("Install with:")
		fmt.Println("  go install github.com/air-verse/air@latest")
		return fmt.Errorf("air not found")
	}

	fmt.Println("Building bin/do...")
	build := exec.Command("go", "build", "-o", "bin/do", "./cmd/do")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("failed to build do: %w", err)
	}

	airArgs := []string{
		"air",
		"-c", "/dev/null",
		"-root", ".",
		"-build.cmd", "./bin/do gen && go build -o ./tmp/main ./cmd/server",
		"-build.bin", "./tmp/main",
		"-build.delay", "100",
		"-build.exclude_dir", "bin,node_modules,tmp,.data",
		"-build.exclude_regex", "_templ.go$|_test.go$|output\\.css$",
		"-build.include_ext", "go,templ,css",
		"-build.kill_delay", "500ms",
		"-build.send_interrupt", "true",
		"-proxy.enabled", "true",
		"-proxy.proxy_port", "8080",
		"-proxy.app_port", "8090",
	}

	env := os.Environ()
	env = append(env, "PORT=8090")

	return syscall.Exec(airPath, airArgs, env)
}
