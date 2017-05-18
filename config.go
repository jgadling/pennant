package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Config struct {
	StorageDriver string       `json:"storage_driver"`
	StatsD        StatsDConfig `json:"statsd_server"`
	HttpPort      int          `json:"http_port"`
	GrpcAddr      string       `json:"grpc_addr"`
	GrpcPort      int          `json:"grpc_port"`
	Consul        ConsulConfig `json:"consul"`
}

func loadConfig(conf_file string) (*Config, error) {
	conf := Config{}
	_, err := os.Stat(conf_file)
	logger.Debug("loading config")
	file, err := ioutil.ReadFile(conf_file)
	if err != nil {
		return &conf, err
	}
	if err = json.Unmarshal(file, &conf); err != nil {
		return &conf, err
	}
	return &conf, nil
}

func (conf *Config) getDriver() (StorageDriver, error) {
	switch conf.StorageDriver {
	case "consul":
		return NewConsulDriver(&conf.Consul)
	}
	return nil, errors.New("invalid driver configuration")
}
