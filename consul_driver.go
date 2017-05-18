package main

import (
	"encoding/json"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

type StorageDriver interface {
	watchForChanges(*FlagCache, uint64)
	loadAllFlags(*FlagCache) (uint64, error)
	createFlag(*Flag) error
	saveFlag(*Flag) error
	deleteFlag(string) error
}

type ConsulConfig struct {
	Protocol    string `json:"protocol"`
	Host        string `json:"host"`
	Scheme      string `json:"scheme"`
	AuthToken   string `json:"auth_token"`
	Prefix      string `json:"prefix"`
	PollTimeout int    `json:"poll_timeout"`
}

type ConsulDriver struct {
	conf   *ConsulConfig
	client *consulapi.Client
}

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

func getConsulConnection(conf *ConsulConfig) (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = conf.Host
	config.Token = conf.AuthToken
	config.WaitTime = time.Duration(conf.PollTimeout) * time.Second
	config.Scheme = conf.Scheme
	consul, _ := consulapi.NewClient(config)
	return consul, nil
}

func (driver *ConsulDriver) watchForChanges(fc *FlagCache, version uint64) {
	for true {
		data, newVersion, err := driver.GetPolicies(version)
		if err != nil {
			// Give consul a few seconds to get its act together
			logger.Warning("Got an error from consul: %v", err)
			time.Sleep(3 * time.Second)
		}
		driver.updateCache(fc, data)
		version = newVersion
	}
}

func (driver *ConsulDriver) loadAllFlags(fc *FlagCache) (uint64, error) {
	// On start, read all policies and add them to our flag cache.
	data, version, err := driver.GetPolicies(0)
	if err != nil {
		return uint64(0), err
	}
	driver.updateCache(fc, data)
	return version, nil
}

func (driver *ConsulDriver) GetPolicies(version uint64) (consulapi.KVPairs, uint64, error) {
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

func (driver *ConsulDriver) updateCache(fc *FlagCache, data consulapi.KVPairs) error {
	// Update all flags from consul.
	foundFlags := make(map[string]bool)
	for _, flagItem := range data {
		logger.Warningf("Flag modifyIndex is %v", flagItem.ModifyIndex)
		flag, err := LoadFlagJson(flagItem.Value)
		foundFlags[flag.Name] = true
		if err != nil {
			// If there's an error in a single flag, log it and move on
			logger.Criticalf("can't parse flag %s", flagItem.Value)
			continue
		}
		oldFlag, err := fc.Get(flag.Name)
		if err == nil && oldFlag.Version == flagItem.ModifyIndex {
			logger.Warningf("Flag %s not changed, skipping update", flag.Name)
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
	logger.Warningf("Found Flags %s", foundFlags)
	cachedFlags := fc.List()
	for idx, _ := range cachedFlags {
		_, ok := foundFlags[idx]
		if !ok {
			logger.Warningf("Deleting cache key %s", idx)
			fc.Delete(idx)
		}
	}

	return nil
}

func (driver *ConsulDriver) saveFlag(f *Flag) error {
	client := driver.client.KV()
	prefix := driver.conf.Prefix
	flagKey := fmt.Sprintf("%s/%s", prefix, f.Name)
	flagJson, _ := json.Marshal(f)
	logger.Criticalf("Key is %s", flagKey)
	logger.Criticalf("Value is %s", flagJson)
	flagVal := &consulapi.KVPair{
		Key:   flagKey,
		Value: flagJson}

	_, err := client.Put(flagVal, nil)
	return err
}

func (driver *ConsulDriver) deleteFlag(flagName string) error {
	client := driver.client.KV()
	prefix := driver.conf.Prefix
	flagKey := fmt.Sprintf("%s/%s", prefix, flagName)
	_, err := client.Delete(flagKey, nil)
	return err
}

func (driver *ConsulDriver) createFlag(f *Flag) error {
	return nil
}
