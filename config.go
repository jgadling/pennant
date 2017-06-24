package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// Config represents the application configuration.
type Config struct {
	StorageDriver string       `json:"storage_driver"`
	StatsD        StatsDConfig `json:"statsd_server"`
	HTTPPort      int          `json:"http_port"`
	GrpcAddr      string       `json:"grpc_addr"`
	GrpcPort      int          `json:"grpc_port"`
	Consul        ConsulConfig `json:"consul"`
}

// Entrypoint to loading server configuration.
func loadConfig(cfg string) (*Config, error) {
	conf := Config{}
	_, err := os.Stat(cfg)
	file, err := ioutil.ReadFile(cfg)
	if err != nil {
		return &conf, err
	}
	if err = json.Unmarshal(file, &conf); err != nil {
		return &conf, err
	}
	return &conf, nil
}

// Load the flag storage driver.
func (conf *Config) getDriver() (StorageDriver, error) {
	switch conf.StorageDriver {
	case "consul":
		return NewConsulDriver(&conf.Consul)
	}
	return nil, errors.New("invalid driver configuration")
}
