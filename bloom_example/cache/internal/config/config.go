package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ConfYaml struct {
	LogOptions *LogOptions
	Overrides  *Overrides

	Server  ServerSection  `yaml:"server"`
	Metrics MetricsSection `yaml:"metrics"`
}

type MetricsSection struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

type ServerSection struct {
	Port    int    `yaml:"port"`
	PidFile string `yaml:"pidFile"`
}

type LogOptions struct {
	Format string
	Level  string
	Log    string
}

func (t *LogOptions) AsString() string {
	var out []string
	out = append(out, "level:'"+t.Level+"'")
	out = append(out, "log:'"+t.Log+"'")
	return strings.Join(out, ",")
}

type Overrides struct {
	Debug bool
	Log   string
}

func (o *Overrides) AsString() string {
	var out []string
	out = append(out, "debug:'"+boolToString(o.Debug)+"'")
	out = append(out, "log:'"+o.Log+"'")
	return strings.Join(out, ",")
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

var defaultConf = []byte(`
log:
  format: "string"
  log: "stdout"
  level: "debug"

server:
  port: 8080
  pidfile: "./cache.pid"

metrics:
  enabled: true
  prefix: "cache"
`)

func LoadConf(confPath string, overrides Overrides) (ConfYaml, error) {
	var conf ConfYaml

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("cache")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var err error
	if confPath != "" {
		if _, err = os.Stat(confPath); err != nil {
			return conf, err
		}

		var content []byte
		if content, err = ioutil.ReadFile(confPath); err != nil {
			return conf, err
		}
		if err = viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return conf, err
		}
	} else {
		viper.AddConfigPath("/etc/cache")
		viper.AddConfigPath("$HOME/.cache")
		viper.AddConfigPath(".")
		viper.SetConfigName("cache")

		if err := viper.ReadInConfig(); err == nil {
			// config found
		} else {
			if err := viper.ReadConfig(bytes.NewBuffer(defaultConf)); err != nil {
				return conf, err
			}
		}
	}

	conf.Overrides = &overrides

	var logOptions LogOptions
	logOptions.Format = viper.GetString("log.format")

	logOptions.Level = viper.GetString("log.level")
	if overrides.Debug {
		logOptions.Level = "debug"
	}

	logOptions.Log = viper.GetString("log.log")
	if len(overrides.Log) > 0 {
		logOptions.Log = overrides.Log
	}

	conf.LogOptions = &logOptions

	conf.Server.Port = viper.GetInt("server.port")
	conf.Server.PidFile = viper.GetString("server.pidFile")

	conf.Metrics.Enabled = viper.GetBool("metrics.enabled")
	conf.Metrics.Prefix = viper.GetString("metrics.prefix")

	return conf, nil
}
