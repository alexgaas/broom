package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

type cmdMgmt struct {
	g *CmdGlobal
}

func (c *cmdMgmt) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "mgmt"
	cmd.Short = "Management commands"
	cmd.Long = `Description:
  Management commands for cache service.
`

	showConfigCmd := cmdShowConfig{g: c.g}
	cmd.AddCommand(showConfigCmd.Command())

	configCmd := cmdConfig{g: c.g}
	cmd.AddCommand(configCmd.Command())

	statusCmd := cmdStatus{g: c.g}
	cmd.AddCommand(statusCmd.Command())

	return cmd
}

// command: "cache mgmt show-config"
type cmdShowConfig struct {
	g *CmdGlobal

	asJson bool
}

func (c *cmdShowConfig) Command() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "show-config"
	cmd.Short = "Show current configuration"
	cmd.Long = "Show current configuration"

	cmd.PersistentFlags().BoolVarP(&c.asJson, "json", "j", false, "output as JSON")

	cmd.RunE = c.Run
	return cmd
}

func (c *cmdShowConfig) Run(cmd *cobra.Command, args []string) error {
	if c.asJson {
		return c.showAsJson()
	}
	return c.showAsText()
}

func (c *cmdShowConfig) showAsJson() error {
	output, err := json.MarshalIndent(c.g.Opts, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (c *cmdShowConfig) showAsText() error {
	fmt.Println("=== Cache Service Configuration ===")
	fmt.Println()

	fmt.Println("[Server]")
	fmt.Printf("  Port:     %d\n", c.g.Opts.Server.Port)
	fmt.Printf("  PID File: %s\n", c.g.Opts.Server.PidFile)
	fmt.Println()

	if c.g.Opts.LogOptions != nil {
		fmt.Println("[Logging]")
		fmt.Printf("  Format: %s\n", c.g.Opts.LogOptions.Format)
		fmt.Printf("  Level:  %s\n", c.g.Opts.LogOptions.Level)
		fmt.Printf("  Output: %s\n", c.g.Opts.LogOptions.Log)
		fmt.Println()
	}

	if c.g.Opts.Overrides != nil {
		fmt.Println("[Overrides]")
		fmt.Printf("  Debug: %t\n", c.g.Opts.Overrides.Debug)
		fmt.Printf("  Log:   %s\n", c.g.Opts.Overrides.Log)
	}

	return nil
}

// command: "cache mgmt config"
type cmdConfig struct {
	g *CmdGlobal

	format string
}

func (c *cmdConfig) Command() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "config"
	cmd.Short = "Output configuration"
	cmd.Long = "Output current configuration in specified format"

	cmd.PersistentFlags().StringVarP(&c.format, "format", "f", "text", "output format: text, json, yaml")

	cmd.RunE = c.Run
	return cmd
}

func (c *cmdConfig) Run(cmd *cobra.Command, args []string) error {
	switch c.format {
	case "json":
		return c.outputJson()
	case "yaml":
		return c.outputYaml()
	default:
		return c.outputText()
	}
}

func (c *cmdConfig) outputJson() error {
	output, err := json.MarshalIndent(c.g.Opts, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (c *cmdConfig) outputYaml() error {
	fmt.Println("log:")
	if c.g.Opts.LogOptions != nil {
		fmt.Printf("  format: \"%s\"\n", c.g.Opts.LogOptions.Format)
		fmt.Printf("  level: \"%s\"\n", c.g.Opts.LogOptions.Level)
		fmt.Printf("  log: \"%s\"\n", c.g.Opts.LogOptions.Log)
	}
	fmt.Println()
	fmt.Println("server:")
	fmt.Printf("  port: %d\n", c.g.Opts.Server.Port)
	fmt.Printf("  pidfile: \"%s\"\n", c.g.Opts.Server.PidFile)
	return nil
}

func (c *cmdConfig) outputText() error {
	fmt.Printf("Server Port:     %d\n", c.g.Opts.Server.Port)
	fmt.Printf("Server PID File: %s\n", c.g.Opts.Server.PidFile)
	if c.g.Opts.LogOptions != nil {
		fmt.Printf("Log Format:      %s\n", c.g.Opts.LogOptions.Format)
		fmt.Printf("Log Level:       %s\n", c.g.Opts.LogOptions.Level)
		fmt.Printf("Log Output:      %s\n", c.g.Opts.LogOptions.Log)
	}
	return nil
}

// command: "cache mgmt status"
type cmdStatus struct {
	g *CmdGlobal

	asJson bool
}

func (c *cmdStatus) Command() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "status"
	cmd.Short = "Show status of detached running services"
	cmd.Long = "Show status of detached running services"

	cmd.PersistentFlags().BoolVarP(&c.asJson, "json", "j", false, "output as JSON")

	cmd.RunE = c.Run
	return cmd
}

func (c *cmdStatus) Run(cmd *cobra.Command, args []string) error {
	pidfile := c.g.Opts.Server.PidFile
	status := c.getServiceStatus(pidfile)

	if c.asJson {
		return c.outputJson(status)
	}
	return c.outputText(status)
}

type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	PID     int    `json:"pid,omitempty"`
	PIDFile string `json:"pid_file"`
	Port    int    `json:"port"`
}

func (c *cmdStatus) getServiceStatus(pidfile string) ServiceStatus {
	status := ServiceStatus{
		Name:    "cache",
		PIDFile: pidfile,
		Port:    c.g.Opts.Server.Port,
		Status:  "stopped",
	}

	content, err := ioutil.ReadFile(pidfile)
	if err != nil {
		if os.IsNotExist(err) {
			status.Status = "stopped (no pidfile)"
		} else {
			status.Status = fmt.Sprintf("error: %s", err)
		}
		return status
	}

	if len(content) == 0 {
		status.Status = "stopped (empty pidfile)"
		return status
	}

	pid, err := strconv.ParseInt(string(content), 10, 64)
	if err != nil {
		status.Status = "error: invalid pid in pidfile"
		return status
	}

	status.PID = int(pid)

	process, err := os.FindProcess(int(pid))
	if err != nil {
		status.Status = "stopped"
		return status
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		status.Status = "stopped (stale pidfile)"
		return status
	}

	status.Status = "running"
	return status
}

func (c *cmdStatus) outputJson(status ServiceStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (c *cmdStatus) outputText(status ServiceStatus) error {
	fmt.Println("=== Detached Services Status ===")
	fmt.Println()
	fmt.Printf("Service:  %s\n", status.Name)
	fmt.Printf("Status:   %s\n", status.Status)
	if status.PID > 0 {
		fmt.Printf("PID:      %d\n", status.PID)
	}
	fmt.Printf("PID File: %s\n", status.PIDFile)
	fmt.Printf("Port:     %d\n", status.Port)
	return nil
}
