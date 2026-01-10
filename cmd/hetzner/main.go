package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	sshHost    string
	sshPort    string
	sshKeyPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hetzner",
		Short: "Hetzner server management CLI",
	}

	// Persistent flags for SSH connection
	rootCmd.PersistentFlags().StringVar(&sshHost, "host", os.Getenv("SSH_HOST"), "SSH host (user@host) or set SSH_HOST env")
	rootCmd.PersistentFlags().StringVar(&sshPort, "port", "22", "SSH port")
	rootCmd.PersistentFlags().StringVar(&sshKeyPath, "key", "", "Path to SSH private key (default: ~/.ssh/id_ed25519)")

	// services command
	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "Manage systemctl services",
	}

	// services list
	var showAll bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List systemctl services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return listServices(sshHost, sshPort, sshKeyPath, showAll)
		},
	}
	listCmd.Flags().BoolVar(&showAll, "all", false, "Show all services including inactive/exited")

	// services delete
	deleteCmd := &cobra.Command{
		Use:   "delete <service-name>",
		Short: "Stop, disable, and delete a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return deleteService(sshHost, sshPort, sshKeyPath, args[0])
		},
	}

	// services create
	var createName, createCmd, createConfig string
	createCmdObj := &cobra.Command{
		Use:   "create",
		Short: "Build, deploy, and create a systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return createService(sshHost, sshPort, sshKeyPath, createName, createCmd, createConfig)
		},
	}
	// Get default name from current directory
	defaultName := ""
	if cwd, err := os.Getwd(); err == nil {
		defaultName = filepath.Base(cwd)
	}
	createCmdObj.Flags().StringVar(&createName, "name", defaultName, "Service name (default: directory name)")
	createCmdObj.Flags().StringVar(&createCmd, "cmd", "cmd/server", "Go package to build")
	createCmdObj.Flags().StringVar(&createConfig, "config", ".env.production", "Config file to deploy")

	servicesCmd.AddCommand(listCmd, deleteCmd, createCmdObj)
	rootCmd.AddCommand(servicesCmd)

	// sites command
	sitesCmd := &cobra.Command{
		Use:   "sites",
		Short: "Manage Caddy sites",
	}

	// sites list
	sitesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List Caddy sites",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return listSites(sshHost, sshPort, sshKeyPath)
		},
	}

	// sites show
	sitesShowCmd := &cobra.Command{
		Use:   "show <domain>",
		Short: "Show config file for a site",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return showSite(sshHost, sshPort, sshKeyPath, args[0])
		},
	}

	// sites delete
	sitesDeleteCmd := &cobra.Command{
		Use:   "delete <domain>",
		Short: "Delete a Caddy site",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return deleteSite(sshHost, sshPort, sshKeyPath, args[0])
		},
	}

	// sites create
	var siteCreateName, siteCreateConfig string
	sitesCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Caddy site from config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sshHost == "" {
				return fmt.Errorf("--host is required or set SSH_HOST env")
			}
			return createSite(sshHost, sshPort, sshKeyPath, siteCreateName, siteCreateConfig)
		},
	}
	sitesCreateCmd.Flags().StringVar(&siteCreateName, "name", defaultName, "Site name (default: directory name)")
	sitesCreateCmd.Flags().StringVar(&siteCreateConfig, "config", ".env.production", "Config file with APP_URL and PORT")

	sitesCmd.AddCommand(sitesListCmd, sitesShowCmd, sitesDeleteCmd, sitesCreateCmd)
	rootCmd.AddCommand(sitesCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type Service struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

func listServices(host, port, keyPath string, all bool) error {
	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	cmd := "systemctl list-units --type=service --no-pager --output=json"
	if all {
		cmd += " --all"
	}

	output, err := runSSHCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	var services []Service
	if err := json.Unmarshal([]byte(output), &services); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	// Filter services
	var filtered []Service
	for _, s := range services {
		// Skip systemd-* services
		if strings.HasPrefix(s.Unit, "systemd-") {
			continue
		}
		// Skip exited services unless --all
		if !all && s.Sub == "exited" {
			continue
		}
		filtered = append(filtered, s)
	}

	// Sort alphabetically by unit name
	slices.SortFunc(filtered, func(a, b Service) int {
		return strings.Compare(a.Unit, b.Unit)
	})

	// Print results
	fmt.Printf("%-40s %-8s %-10s %s\n", "UNIT", "ACTIVE", "SUB", "DESCRIPTION")
	for _, s := range filtered {
		fmt.Printf("%-40s %-8s %-10s %s\n", s.Unit, s.Active, s.Sub, s.Description)
	}

	return nil
}

func deleteService(host, port, keyPath, serviceName string) error {
	// Ensure service name ends with .service
	if !strings.HasSuffix(serviceName, ".service") {
		serviceName = serviceName + ".service"
	}

	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	// Find and show the service file
	servicePath := "/etc/systemd/system/" + serviceName
	config, err := runSSHCommand(client, "cat "+servicePath)
	if err != nil {
		return fmt.Errorf("service file not found at %s: %w", servicePath, err)
	}

	fmt.Printf("Service file: %s\n", servicePath)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(config)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
	fmt.Printf("This will STOP, DISABLE, and DELETE %s\n", serviceName)
	fmt.Print("Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	if strings.TrimSpace(response) != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

	// Stop the service
	fmt.Printf("Stopping %s...\n", serviceName)
	if _, err := runSSHCommand(client, "systemctl stop "+serviceName); err != nil {
		fmt.Printf("Warning: stop failed (may already be stopped): %v\n", err)
	}

	// Disable the service
	fmt.Printf("Disabling %s...\n", serviceName)
	if _, err := runSSHCommand(client, "systemctl disable "+serviceName); err != nil {
		fmt.Printf("Warning: disable failed: %v\n", err)
	}

	// Delete the service file
	fmt.Printf("Deleting %s...\n", servicePath)
	if _, err := runSSHCommand(client, "rm "+servicePath); err != nil {
		return fmt.Errorf("delete service file: %w", err)
	}

	// Reload systemd
	fmt.Println("Reloading systemd...")
	if _, err := runSSHCommand(client, "systemctl daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	fmt.Printf("Service %s deleted successfully.\n", serviceName)
	return nil
}

func createService(host, port, keyPath, name, cmdPath, configPath string) error {
	if name == "" {
		return fmt.Errorf("service name is required")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// Cross-compile for Linux
	fmt.Printf("Building %s for linux/amd64...\n", cmdPath)
	tmpBinary := filepath.Join(os.TempDir(), name+"-linux-amd64")
	buildCmd := exec.Command("go", "build", "-o", tmpBinary, "./"+cmdPath)
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	defer os.Remove(tmpBinary)

	// Connect to server
	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	// Create directory
	srvPath := "/srv/" + name
	fmt.Printf("Creating %s...\n", srvPath)
	if _, err := runSSHCommand(client, "mkdir -p "+srvPath); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Upload binary
	fmt.Printf("Uploading binary to %s/app...\n", srvPath)
	if err := scpFile(client, tmpBinary, srvPath+"/app", 0755); err != nil {
		return fmt.Errorf("upload binary: %w", err)
	}

	// Upload config
	fmt.Printf("Uploading config to %s/.env...\n", srvPath)
	if err := scpFile(client, configPath, srvPath+"/.env", 0644); err != nil {
		return fmt.Errorf("upload config: %w", err)
	}

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s/app
Restart=always
RestartSec=5
EnvironmentFile=%s/.env

[Install]
WantedBy=multi-user.target
`, name, srvPath, srvPath, srvPath)

	servicePath := "/etc/systemd/system/" + name + ".service"
	fmt.Printf("Creating systemd service %s...\n", servicePath)

	// Write service file via SSH
	escaped := strings.ReplaceAll(serviceContent, "'", "'\\''")
	if _, err := runSSHCommand(client, fmt.Sprintf("echo '%s' > %s", escaped, servicePath)); err != nil {
		return fmt.Errorf("create service file: %w", err)
	}

	// Reload systemd
	fmt.Println("Reloading systemd...")
	if _, err := runSSHCommand(client, "systemctl daemon-reload"); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	// Enable and start service
	fmt.Printf("Enabling %s...\n", name)
	if _, err := runSSHCommand(client, "systemctl enable "+name); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	fmt.Printf("Starting %s...\n", name)
	if _, err := runSSHCommand(client, "systemctl start "+name); err != nil {
		return fmt.Errorf("start service: %w", err)
	}

	// Check status
	status, _ := runSSHCommand(client, "systemctl is-active "+name)
	fmt.Printf("\nService %s created and started (%s)\n", name, strings.TrimSpace(status))
	fmt.Printf("  Binary: %s/app\n", srvPath)
	fmt.Printf("  Config: %s/.env\n", srvPath)
	fmt.Printf("  Service: %s\n", servicePath)

	return nil
}

func scpFile(client *ssh.Client, localPath, remotePath string, mode os.FileMode) error {
	// Open local file
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Start scp on remote
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", mode, stat.Size(), filepath.Base(remotePath))
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	return session.Run("scp -t " + remotePath)
}

type CaddySite struct {
	Domain       string
	Root         string
	ReverseProxy string
}

func listSites(host, port, keyPath string) error {
	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	// Query Caddy API
	output, err := runSSHCommand(client, "curl -s http://localhost:2019/config/apps/http/servers")
	if err != nil {
		return fmt.Errorf("caddy API: %w", err)
	}

	sites, err := parseCaddyAPI(output)
	if err != nil {
		return fmt.Errorf("parse caddy config: %w", err)
	}

	// Sort alphabetically by domain
	slices.SortFunc(sites, func(a, b CaddySite) int {
		return strings.Compare(a.Domain, b.Domain)
	})

	// Print results
	fmt.Printf("%-35s %-30s %s\n", "DOMAIN", "ROOT", "REVERSE_PROXY")
	for _, s := range sites {
		// Skip the default :80 fallback
		if s.Domain == ":80" && s.Root == "/usr/share/caddy" {
			continue
		}
		root := s.Root
		if root == "" {
			root = "-"
		}
		proxy := s.ReverseProxy
		if proxy == "" {
			proxy = "-"
		}
		fmt.Printf("%-35s %-30s %s\n", s.Domain, root, proxy)
	}

	return nil
}

func findSiteConfigFile(client *ssh.Client, domain string) (string, string, error) {
	// Check common locations for site config files
	paths := []string{
		"/etc/caddy/sites/" + domain,
		"/etc/caddy/sites.d/" + domain,
		"/etc/caddy/conf.d/" + domain,
		"/etc/caddy/sites/" + domain + ".caddy",
		"/etc/caddy/sites.d/" + domain + ".caddy",
		"/etc/caddy/conf.d/" + domain + ".caddy",
	}

	for _, path := range paths {
		content, err := runSSHCommand(client, "cat "+path+" 2>/dev/null")
		if err == nil && strings.TrimSpace(content) != "" {
			return path, content, nil
		}
	}

	// Try to find by grepping for the domain
	files, err := runSSHCommand(client, "grep -l '"+domain+"' /etc/caddy/sites/* /etc/caddy/sites.d/* /etc/caddy/conf.d/* 2>/dev/null | head -1")
	if err == nil && strings.TrimSpace(files) != "" {
		path := strings.TrimSpace(strings.Split(files, "\n")[0])
		content, err := runSSHCommand(client, "cat "+path)
		if err == nil {
			return path, content, nil
		}
	}

	return "", "", fmt.Errorf("config file not found for domain: %s", domain)
}

func showSite(host, port, keyPath, domain string) error {
	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	path, content, err := findSiteConfigFile(client, domain)
	if err != nil {
		return err
	}

	fmt.Printf("Config file: %s\n", path)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Print(content)
	if !strings.HasSuffix(content, "\n") {
		fmt.Println()
	}
	fmt.Println(strings.Repeat("-", 60))

	return nil
}

func deleteSite(host, port, keyPath, domain string) error {
	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	path, content, err := findSiteConfigFile(client, domain)
	if err != nil {
		return err
	}

	fmt.Printf("Config file: %s\n", path)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Print(content)
	if !strings.HasSuffix(content, "\n") {
		fmt.Println()
	}
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
	fmt.Printf("This will DELETE the site %s and reload Caddy\n", domain)
	fmt.Print("Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	if strings.TrimSpace(response) != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

	// Delete the config file
	fmt.Printf("Deleting %s...\n", path)
	if _, err := runSSHCommand(client, "rm "+path); err != nil {
		return fmt.Errorf("delete config file: %w", err)
	}

	// Reload Caddy
	fmt.Println("Reloading Caddy...")
	if _, err := runSSHCommand(client, "systemctl reload caddy"); err != nil {
		return fmt.Errorf("reload caddy: %w", err)
	}

	fmt.Printf("Site %s deleted successfully.\n", domain)
	return nil
}

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parse KEY=VALUE
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			// Remove inline comments
			if commentIdx := strings.Index(value, " #"); commentIdx > 0 {
				value = strings.TrimSpace(value[:commentIdx])
			}
			env[key] = value
		}
	}
	return env, scanner.Err()
}

func createSite(host, port, keyPath, name, configPath string) error {
	if name == "" {
		return fmt.Errorf("site name is required")
	}

	// Parse config file for APP_URL and PORT
	env, err := parseEnvFile(configPath)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	appURL := env["APP_URL"]
	appPort := env["PORT"]

	if appURL == "" {
		return fmt.Errorf("APP_URL not found in %s", configPath)
	}
	if appPort == "" {
		return fmt.Errorf("PORT not found in %s", configPath)
	}

	// Extract domain from APP_URL (remove protocol)
	domain := appURL
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")

	fmt.Printf("Creating Caddy site:\n")
	fmt.Printf("  Domain: %s\n", domain)
	fmt.Printf("  Proxy:  localhost:%s\n", appPort)

	// Connect to server
	client, err := sshConnect(host, port, keyPath)
	if err != nil {
		return fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()

	// Create Caddy config
	caddyConfig := fmt.Sprintf(`%s {
	reverse_proxy localhost:%s
}
`, domain, appPort)

	// Ensure sites directory exists
	if _, err := runSSHCommand(client, "mkdir -p /etc/caddy/sites"); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Write config file
	configFilePath := "/etc/caddy/sites/" + name
	fmt.Printf("Creating %s...\n", configFilePath)
	escaped := strings.ReplaceAll(caddyConfig, "'", "'\\''")
	if _, err := runSSHCommand(client, fmt.Sprintf("echo '%s' > %s", escaped, configFilePath)); err != nil {
		return fmt.Errorf("create config file: %w", err)
	}

	// Check if import exists in Caddyfile
	caddyfile, _ := runSSHCommand(client, "cat /etc/caddy/Caddyfile")
	hasImport := strings.Contains(caddyfile, "import /etc/caddy/sites/*") ||
		strings.Contains(caddyfile, "import sites/*")
	if !hasImport {
		fmt.Println("Adding import directive to Caddyfile...")
		if _, err := runSSHCommand(client, `sed -i '1i import /etc/caddy/sites/*' /etc/caddy/Caddyfile`); err != nil {
			return fmt.Errorf("update Caddyfile: %w", err)
		}
	}

	// Reload Caddy
	fmt.Println("Reloading Caddy...")
	if output, err := runSSHCommand(client, "systemctl reload caddy 2>&1"); err != nil {
		// Try to get more details
		journalOutput, _ := runSSHCommand(client, "journalctl -u caddy -n 10 --no-pager 2>&1")
		return fmt.Errorf("reload caddy: %w\n%s\n%s", err, output, journalOutput)
	}

	fmt.Printf("\nSite %s created successfully.\n", domain)
	fmt.Printf("  Config: %s\n", configFilePath)
	fmt.Printf("  URL: https://%s\n", domain)

	return nil
}

func parseCaddyAPI(content string) ([]CaddySite, error) {
	var servers map[string]struct {
		Listen []string `json:"listen"`
		Routes []struct {
			Match []struct {
				Host []string `json:"host"`
			} `json:"match"`
			Handle []struct {
				Handler  string `json:"handler"`
				Root     string `json:"root"`
				Upstreams []struct {
					Dial string `json:"dial"`
				} `json:"upstreams"`
				Routes []struct {
					Handle []struct {
						Handler  string `json:"handler"`
						Root     string `json:"root"`
						Upstreams []struct {
							Dial string `json:"dial"`
						} `json:"upstreams"`
					} `json:"handle"`
				} `json:"routes"`
			} `json:"handle"`
		} `json:"routes"`
	}

	if err := json.Unmarshal([]byte(content), &servers); err != nil {
		return nil, err
	}

	var sites []CaddySite
	for _, server := range servers {
		for _, route := range server.Routes {
			site := CaddySite{}

			// Get domain from match
			if len(route.Match) > 0 && len(route.Match[0].Host) > 0 {
				site.Domain = route.Match[0].Host[0]
			}

			// Get root and reverse_proxy from handlers
			for _, handle := range route.Handle {
				if handle.Handler == "file_server" && handle.Root != "" {
					site.Root = handle.Root
				}
				if handle.Handler == "reverse_proxy" && len(handle.Upstreams) > 0 {
					site.ReverseProxy = handle.Upstreams[0].Dial
				}
				// Check nested routes (subroutes)
				for _, subroute := range handle.Routes {
					for _, subhandle := range subroute.Handle {
						if subhandle.Handler == "file_server" && subhandle.Root != "" {
							site.Root = subhandle.Root
						}
						if subhandle.Handler == "reverse_proxy" && len(subhandle.Upstreams) > 0 {
							site.ReverseProxy = subhandle.Upstreams[0].Dial
						}
					}
				}
			}

			if site.Domain != "" {
				sites = append(sites, site)
			}
		}
	}

	return sites, nil
}

func runSSHCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

func sshConnect(host, port, keyPath string) (*ssh.Client, error) {
	authMethods, err := getAuthMethods(keyPath)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            parseUser(host),
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%s", parseHost(host), port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	return client, nil
}

func getAuthMethods(keyPath string) ([]ssh.AuthMethod, error) {
	// Try ssh-agent first
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			keys, err := agentClient.List()
			if err == nil && len(keys) > 0 {
				return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, nil
			}

			// No keys in agent, try to add one
			if err := runSSHAdd(); err != nil {
				return nil, fmt.Errorf("ssh-add failed: %w", err)
			}

			// Reconnect and try again
			conn, err = net.Dial("unix", sock)
			if err == nil {
				agentClient = agent.NewClient(conn)
				return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, nil
			}
		}
	}

	// Fall back to key file
	var key []byte
	var err error

	if keyPath != "" {
		key, err = os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read key %s: %w", keyPath, err)
		}
	} else {
		key, _, err = findSSHKey()
		if err != nil {
			return nil, err
		}
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse key (use ssh-add to load passphrase-protected keys): %w", err)
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

func runSSHAdd() error {
	fmt.Println("No keys in ssh-agent, running ssh-add...")
	cmd := exec.Command("ssh-add")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findSSHKey() ([]byte, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("get home dir: %w", err)
	}

	keyNames := []string{"id_ed25519", "id_rsa", "id_ecdsa", "id_dsa"}
	for _, name := range keyNames {
		path := filepath.Join(home, ".ssh", name)
		key, err := os.ReadFile(path)
		if err == nil {
			return key, path, nil
		}
	}

	return nil, "", fmt.Errorf("no SSH key found in ~/.ssh (tried: %v)", keyNames)
}

func parseUser(host string) string {
	for i, c := range host {
		if c == '@' {
			return host[:i]
		}
	}
	return "root"
}

func parseHost(host string) string {
	for i, c := range host {
		if c == '@' {
			return host[i+1:]
		}
	}
	return host
}
