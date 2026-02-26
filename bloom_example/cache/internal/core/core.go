package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"

	"cache/internal/api"
	"cache/internal/bch"
	"cache/internal/config"
	"cache/internal/log"
	"cache/internal/metrics"

	"golang.org/x/sys/unix"
)

type Core struct {
	Opts *config.ConfYaml
	Log  *log.TLog

	httpapi *api.Api
}

type Overrides struct {
	Detach bool `json:"detach"`
}

func CreateCore(opts *config.ConfYaml, logger *log.TLog) (*Core, error) {
	var core Core
	core.Opts = opts
	core.Log = logger
	return &core, nil
}

func (c *Overrides) AsString() string {
	return fmt.Sprintf("detach:'%t'", c.Detach)
}

func (c *Core) Daemon() error {
	if os.Getppid() != 1 {
		// I am the parent, spawn child to run as daemon
		binary, err := exec.LookPath(os.Args[0])
		if err != nil {
			c.Log.Error(fmt.Sprintf("failed to lookup binary: %s", err))
			return err
		}

		_, err = os.StartProcess(binary, os.Args, &os.ProcAttr{
			Dir:   "",
			Env:   os.Environ(),
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Sys:   nil,
		})
		if err != nil {
			c.Log.Error(fmt.Sprintf("failed to start process: %s", err))
			return err
		}

		// Parent exits, child continues
		os.Exit(0)
	} else {
		// I am the child (daemon), start new session and detach from terminal
		_, err := syscall.Setsid()
		if err != nil {
			c.Log.Error(fmt.Sprintf("failed to create new session: %s", err))
			return err
		}

		file, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if err != nil {
			c.Log.Error(fmt.Sprintf("failed to open /dev/null: %s", err))
			return err
		}

		unix.Dup2(int(file.Fd()), int(os.Stdin.Fd()))
		unix.Dup2(int(file.Fd()), int(os.Stdout.Fd()))
		unix.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
		file.Close()

		// Write pidfile
		pidfile := c.Opts.Server.PidFile
		pid := fmt.Sprintf("%d", os.Getpid())
		if err = ioutil.WriteFile(pidfile, []byte(pid), 0644); err != nil {
			c.Log.Error(fmt.Sprintf("error writing pidfile '%s': %s", pidfile, err))
			return err
		}
	}

	return nil
}

func (c *Core) IfProcessRun() (bool, int, *os.Process, error) {
	var pid int64

	pidfile := c.Opts.Server.PidFile
	content, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return false, int(pid), nil, err
	}

	if len(content) == 0 {
		return false, int(pid), nil, errors.New("empty pidfile")
	}

	spid := string(content)
	if pid, err = strconv.ParseInt(spid, 10, 64); err != nil {
		return false, int(pid), nil, err
	}

	p, err := os.FindProcess(int(pid))
	if err != nil {
		return false, int(pid), nil, err
	}

	if err = p.Signal(syscall.Signal(0)); err != nil {
		return false, int(pid), p, errors.New("process not running")
	}

	return true, int(pid), p, nil
}

func (c *Core) StopCore(overrides *Overrides) error {
	if overrides.Detach {
		run, pid, p, err := c.IfProcessRun()
		if err != nil {
			c.Log.Error(fmt.Sprintf("error detecting run process, pid:%d, err:%s", pid, err))
			return err
		}

		if !run {
			err = errors.New("no running process detected")
			c.Log.Error(fmt.Sprintf("no detecting run process, pid:%d", pid))
			return err
		}

		if err = p.Signal(syscall.SIGTERM); err != nil {
			c.Log.Error(fmt.Sprintf("error signalling pid:%d, err:%s", pid, err))
			return err
		}

		c.Log.Info(fmt.Sprintf("sent SIGTERM to pid:%d", pid))
	}

	return nil
}

func (c *Core) StartCore(overrides *Overrides) error {
	if overrides.Detach {
		// Check if process is already running
		if os.Getppid() != 1 {
			run, pid, _, err := c.IfProcessRun()
			if err == nil && run {
				c.Log.Error(fmt.Sprintf("process already running on pid:%d", pid))
				return errors.New("process already running")
			}
		}

		// Daemonize
		if err := c.Daemon(); err != nil {
			c.Log.Error(fmt.Sprintf("error detaching process: %s", err))
			return err
		}
	}

	c.Log.Debug(fmt.Sprintf("starting core, overrides:%s", overrides.AsString()))

	// Initialize bloom cache
	bloomCache, err := bch.New(bch.DefaultConfig())
	if err != nil {
		c.Log.Error(fmt.Sprintf("error creating bloom cache: %s", err))
		return err
	}
	c.Log.Info("bloom cache initialized")

	var appMetrics *metrics.Metrics
	if c.Opts.Metrics.Enabled {
		prefix := c.Opts.Metrics.Prefix
		if prefix == "" {
			prefix = "cache"
		}
		appMetrics = metrics.New(prefix)
		appMetrics.RegisterBloomGauges(bloomCache)
		c.Log.Info("metrics initialized with prefix: " + prefix)
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	httpApi, _ := api.CreateApi(c.Opts, c.Log)
	httpApi.Core = c
	httpApi.BloomCache = bloomCache
	httpApi.Metrics = appMetrics
	c.httpapi = httpApi

	go httpApi.Apiloop(&waitGroup)

	waitGroup.Wait()

	return nil
}
