package cmd

import (
	"fmt"

	"cache/internal/core"

	"github.com/spf13/cobra"
)

type cmdServer struct {
	g *CmdGlobal

	debug bool

	core *core.Core
}

func (c *cmdServer) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "server"
	cmd.Short = "Controlling server process"
	cmd.Long = `Description:
  cache service is started as a process under systemd control or as
  stand-alone server process.
`
	cmd.PersistentFlags().BoolVarP(&c.debug, "debug", "d", false, "debug output")

	serverStartCmd := cmdServerStart{g: c.g}
	serverStartCmd.s = c
	cmd.AddCommand(serverStartCmd.Command())

	serverStopCmd := cmdServerStop{g: c.g}
	serverStopCmd.s = c
	cmd.AddCommand(serverStopCmd.Command())

	return cmd
}

func (c *cmdServer) InitCommand(cmd *cobra.Command, args []string) (string, error) {
	item := ""
	if len(args) > 0 {
		item = args[0]
	}
	return item, nil
}

// command: "cache server start"
type cmdServerStart struct {
	g *CmdGlobal
	s *cmdServer

	detach bool
}

func (c *cmdServerStart) Command() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "start"
	cmd.Short = "Starting server"
	cmd.Long = "Starting server"

	cmd.PersistentFlags().BoolVarP(&c.detach, "detach", "D", false, "detach process")

	cmd.RunE = c.Run
	return cmd
}

func (c *cmdServerStart) Run(cmd *cobra.Command, args []string) error {
	C, err := core.CreateCore(c.g.Opts, c.g.Log)
	if err != nil {
		c.g.Log.Error(fmt.Sprintf("error creating core: %s", err))
		return err
	}

	var overrides core.Overrides
	overrides.Detach = c.detach
	if err := C.StartCore(&overrides); err != nil {
		c.g.Log.Error(fmt.Sprintf("error starting core: %s", err))
		return err
	}

	return nil
}

// command: "cache server stop"
type cmdServerStop struct {
	g *CmdGlobal
	s *cmdServer

	detach bool
}

func (c *cmdServerStop) Command() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "stop"
	cmd.Short = "Stopping server"
	cmd.Long = "Stopping server"

	cmd.PersistentFlags().BoolVarP(&c.detach, "detach", "D", false, "detach process")

	cmd.RunE = c.Run
	return cmd
}

func (c *cmdServerStop) Run(cmd *cobra.Command, args []string) error {
	C, err := core.CreateCore(c.g.Opts, c.g.Log)
	if err != nil {
		c.g.Log.Error(fmt.Sprintf("error creating core: %s", err))
		return err
	}

	var overrides core.Overrides
	overrides.Detach = c.detach
	if err := C.StopCore(&overrides); err != nil {
		c.g.Log.Error(fmt.Sprintf("error stopping core: %s", err))
		return err
	}

	return nil
}
