// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
)

const defaultRequestTimeout = 15

var defaultConfig = &Config{
	RpcUrl:         "https://soroban-testnet.stellar.org",
	Network:        NetworkTestnet,
	SimulatorPath:  "",
	LogLevel:       "info",
	CachePath:      filepath.Join(os.ExpandEnv("$HOME"), ".erst", "cache"),
	RequestTimeout: defaultRequestTimeout,
}

func DefaultConfig() *Config {
	return &Config{
		RpcUrl:         defaultConfig.RpcUrl,
		Network:        defaultConfig.Network,
		SimulatorPath:  defaultConfig.SimulatorPath,
		LogLevel:       defaultConfig.LogLevel,
		CachePath:      defaultConfig.CachePath,
		RequestTimeout: defaultConfig.RequestTimeout,
	}
}

func NewConfig(rpcUrl string, network Network) *Config {
	return &Config{
		RpcUrl:        rpcUrl,
		Network:       network,
		SimulatorPath: defaultConfig.SimulatorPath,
		LogLevel:      defaultConfig.LogLevel,
		CachePath:     defaultConfig.CachePath,
	}
}

func (c *Config) WithSimulatorPath(path string) *Config {
	c.SimulatorPath = path
	return c
}

func (c *Config) WithLogLevel(level string) *Config {
	c.LogLevel = level
	return c
}

func (c *Config) WithCachePath(path string) *Config {
	c.CachePath = path
	return c
}

func (c *Config) WithRequestTimeout(seconds int) *Config {
	c.RequestTimeout = seconds
	return c
}
