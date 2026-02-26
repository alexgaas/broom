package cmd

import (
	"errors"
	"fmt"
	"os"

	"cache/internal/config"
	"cache/internal/core"
	"cache/internal/log"

	"github.com/spf13/cobra"
)

var (
	configFile string
	debug      bool
	logFile    string

	g CmdGlobal

	InitError error

	rootCmd = &cobra.Command{
		SilenceUsage: true,
		Use:          "cache",
		Short:        "cache service application",
		Long:         `Cache service with hello world endpoint`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if InitError != nil {
				os.Exit(1)
			}

			var overrides config.Overrides
			overrides.Debug = debug
			overrides.Log = logFile
			if g.Log != nil {
				g.Log.UpdateOverrides(overrides)
			}
		},
	}
)

type CmdGlobal struct {
	Cmd  *cobra.Command
	Opts *config.ConfYaml
	Log  *log.TLog
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(InitConfig)

	g = CmdGlobal{Cmd: rootCmd, Opts: &config.ConfYaml{}}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "C", "",
		"configuration file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d",
		false, "debug output, default: 'false'")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log", "", "",
		"log file, default: 'stdout'")

	C, _ := core.CreateCore(g.Opts, g.Log)

	serverCmd := cmdServer{g: &g}
	serverCmd.core = C
	rootCmd.AddCommand(serverCmd.Command())

	mgmtCmd := cmdMgmt{g: &g}
	rootCmd.AddCommand(mgmtCmd.Command())
}

func InitConfig() {
	var err error
	var overrides config.Overrides
	overrides.Debug = debug
	overrides.Log = logFile

	*g.Opts, InitError = config.LoadConf(configFile, overrides)

	if InitError != nil {
		fmt.Fprintf(os.Stderr, "config error: %s\n", InitError)
		os.Exit(1)
	}

	if g.Log, err = log.CreateLog(g.Opts); err != nil {
		InitError = errors.New(fmt.Sprintf("logfile initialization error: %s", err))
		os.Exit(1)
	}
}
