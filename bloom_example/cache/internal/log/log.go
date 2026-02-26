package log

import (
	"errors"
	"fmt"
	"os"

	"cache/internal/config"

	"github.com/sirupsen/logrus"
)

const (
	LOGTYPE_STDOUT = 1
	LOGTYPE_FILE   = 2
)

type TLog struct {
	LogType    int
	Log        *logrus.Logger
	LogOptions *config.LogOptions
	File       *os.File
}

func (log *TLog) Debug(str string) {
	if log.LogType == LOGTYPE_STDOUT || log.LogType == LOGTYPE_FILE {
		if log.Log != nil {
			log.Log.Debug(str)
		}
	}
}

func (log *TLog) Info(str string) {
	if log.LogType == LOGTYPE_STDOUT || log.LogType == LOGTYPE_FILE {
		if log.Log != nil {
			log.Log.Info(str)
		}
	}
}

func (log *TLog) Error(str string) {
	if log.LogType == LOGTYPE_STDOUT || log.LogType == LOGTYPE_FILE {
		if log.Log != nil {
			log.Log.Error(str)
		}
	}
}

func (log *TLog) SetLogLevel(level string) error {
	if log.Log == nil {
		return errors.New("log not initialized")
	}
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	log.Log.Level = l
	return nil
}

func (log *TLog) SetLogOut(out string) error {
	if log.Log == nil {
		return errors.New("log not initialized")
	}

	if log.LogType == LOGTYPE_STDOUT || log.LogType == LOGTYPE_FILE {
		switch out {
		case "stdout":
			log.Log.Out = os.Stdout
		case "stderr":
			log.Log.Out = os.Stderr
		default:
			f, err := os.OpenFile(log.LogOptions.Log,
				os.O_RDWR|os.O_CREATE|os.O_APPEND,
				0644)
			if err != nil {
				return err
			}
			log.Log.Out = f
			log.File = f

			logrus.SetOutput(log.File)
			logrus.SetLevel(logrus.DebugLevel)
		}
	}
	return nil
}

func (log *TLog) UpdateLog() error {
	var err error
	if log.LogOptions == nil {
		return errors.New("log file not found")
	}
	if log.LogType == LOGTYPE_STDOUT || log.LogType == LOGTYPE_FILE {
		if err = log.SetLogLevel(log.LogOptions.Level); err != nil {
			return err
		}
		if err = log.SetLogOut(log.LogOptions.Log); err != nil {
			return err
		}
	}
	return err
}

func (log *TLog) UpdateOverrides(overrides config.Overrides) error {
	if log.LogOptions == nil {
		return errors.New("log file not found")
	}
	if log.LogType == LOGTYPE_STDOUT || log.LogType == LOGTYPE_FILE {
		if overrides.Debug {
			log.LogOptions.Level = "debug"
		}
		if len(overrides.Log) > 0 {
			log.LogOptions.Log = overrides.Log
		}
	}
	return log.UpdateLog()
}

func CreateLog(opts *config.ConfYaml) (*TLog, error) {
	var err error

	var Log TLog
	Log.Log = logrus.New()

	logOptions := opts.LogOptions

	Log.LogType = LOGTYPE_STDOUT
	if logOptions.Log != "stdout" {
		Log.LogType = LOGTYPE_FILE
	}

	Log.LogOptions = logOptions
	Log.Log.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006/01/02 - 15:04:05.000",
		FullTimestamp:   true,
	}

	if err = Log.UpdateLog(); err != nil {
		return nil, fmt.Errorf("log error: %s", err.Error())
	}

	return &Log, nil
}
