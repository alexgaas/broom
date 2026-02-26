package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DryRun    bool
	CacheHost string
}

func LoadConf() (*Config, error) {
	dryRun := os.Getenv("DRY_RUN")
	dry, err := strconv.ParseBool(dryRun)
	if err != nil {
		return nil, fmt.Errorf("%s is not set", "DRY_RUN")
	}

	cacheHost := os.Getenv("CACHE_HOST")
	if cacheHost == "" && !dry {
		return nil, fmt.Errorf("%s is not set", "CACHE_HOST")
	}

	return &Config{
		DryRun:    dry,
		CacheHost: cacheHost,
	}, nil
}
