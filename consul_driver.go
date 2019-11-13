package main

import (
	"encoding/json"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

// ConsulConfig is the consul configuration definition, used by config module.
type ConsulConfig struct {
	Protocol    string `json:"protocol"`
	Host        string `json:"host"`
	Scheme      string `json:"scheme"`
	AuthToken   string `json:"auth_token"`
	Prefix      string `json:"prefix"`
	PollTimeout int    `json:"poll_timeout"`
}

// ConsulDriver is a container for consul driver config
type ConsulDriver struct {
	conf   *ConsulConfig
	client *consulapi.Client
}

// NewConsulDriver returns a StorageDriver interface'd object.
func NewConsulDriver(conf *ConsulConfig) (*ConsulDriver, error) {
	driver := &ConsulDriver{
		conf:   conf,
		client: nil,
	}
	client, err := getConsulConnection(conf)
	if err != nil {
		return driver, nil
	}
	driver.client = client
	return driver, nil
}

// Connect to consul
func getConsulConnection(conf *ConsulConfig) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = conf.Host
	config.Token = conf.AuthToken
	config.WaitTime = time.Duration(conf.PollTimeout) * time.Second
	config.Scheme = conf.Scheme
	consul, _ := consulapi.NewClient(config)
	return consul, nil
}

// Main consul-watch loop so we can update our cache when keys change.
func (driver *ConsulDriver) watchForChanges(fc *FlagCache, version uint64) {
	for true {
		data, newVersion, err := driver.GetFlags(version)
		if err != nil {
			// Give consul a few seconds to get its act together
			logger.Warning("Got an error from consul: %v", err)
			time.Sleep(3 * time.Second)
		}
		driver.updateCache(fc, data)
		version = newVersion
	}
}

// On start, read all policies and add them to our flag cache.
func (driver *ConsulDriver) loadAllFlags(fc *FlagCache) (uint64, error) {
	data, version, err := driver.GetFlags(0)
	if err != nil {
		return uint64(0), err
	}
	driver.updateCache(fc, data)
	return version, nil
}

// GetFlags reads all flags from Consul
func (driver *ConsulDriver) GetFlags(version uint64) (consulapi.KVPairs, uint64, error) {
	newVersion := uint64(0)
	client := driver.client.KV()
	prefix := driver.conf.Prefix
	data, meta, err := client.List(prefix, &consulapi.QueryOptions{
		RequireConsistent: true,
		WaitIndex:         version,
	})
	if err != nil {
		return consulapi.KVPairs{}, newVersion, err
	}
	newVersion = meta.LastIndex
	return data, newVersion, nil
}

// Given a list of flag definitions from consul, update our cache as necessary
func (driver *ConsulDriver) updateCache(fc *FlagCache, data consulapi.KVPairs) error {
	foundFlags := make(map[string]bool)
	for _, flagItem := range data {
		flag, err := LoadFlagJSON(flagItem.Value)
		foundFlags[flag.Name] = true
		if err != nil {
			// If there's an error in a single flag, log it and move on
			logger.Criticalf("can't parse flag %s", flagItem.Value)
			continue
		}
		oldFlag, err := fc.Get(flag.Name)
		if err == nil && oldFlag.Version == flagItem.ModifyIndex {
			logger.Debugf("Flag %s not changed, skipping update", flag.Name)
			continue

		}
		flag.Version = flagItem.ModifyIndex
		flag.Parse()
		if err = fc.Upsert(flag); err != nil {
			logger.Criticalf("%v", err)
			continue
		}
	}
	// Remove any flags that were in the cache and no longer in consul
	cachedFlags := fc.List()
	for idx := range cachedFlags {
		_, ok := foundFlags[idx]
		if !ok {
			logger.Debugf("Deleting cache key %s", idx)
			fc.Delete(idx)
		}
	}

	return nil
}

// Write a new flag to consul
func (driver *ConsulDriver) saveFlag(f *Flag) error {
	client := driver.client.KV()
	prefix := driver.conf.Prefix
	flagKey := fmt.Sprintf("%s/%s", prefix, f.Name)
	flagJSON, _ := json.Marshal(f)
	flagVal := &consulapi.KVPair{
		Key:   flagKey,
		Value: flagJSON}

	_, err := client.Put(flagVal, nil)
	return err
}

// Delete a flag from consul
func (driver *ConsulDriver) deleteFlag(flagName string) error {
	client := driver.client.KV()
	prefix := driver.conf.Prefix
	flagKey := fmt.Sprintf("%s/%s", prefix, flagName)
	_, err := client.Delete(flagKey, nil)
	return err
}
