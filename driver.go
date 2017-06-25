package main

import (
	"time"
)

// Storage driver interface
type StorageDriver interface {
	watchForChanges(*FlagCache, uint64)
	loadAllFlags(*FlagCache) (uint64, error)
	saveFlag(*Flag) error
	deleteFlag(string) error
}

// Container for memory driver config
type MemoryDriver struct {
	fc *FlagCache
}

// Return a StorageDriver interface'd object.
func NewMemoryDriver() (*MemoryDriver, error) {
	driver := &MemoryDriver{}
	return driver, nil
}

// Dummy watch loop
func (driver *MemoryDriver) watchForChanges(fc *FlagCache, version uint64) {
	for true {
		time.Sleep(3 * time.Second)
	}
}

// No policies to load on start, since we don't have durable storage.
func (driver *MemoryDriver) loadAllFlags(fc *FlagCache) (uint64, error) {
	driver.fc = fc
	return 1, nil
}

// Update a flag in the in-mem cache
func (driver *MemoryDriver) saveFlag(f *Flag) error {
	if err := driver.fc.Upsert(f); err != nil {
		return err
	}
	return nil
}

// Delete a flag from the in-mem cache
func (driver *MemoryDriver) deleteFlag(flagName string) error {
	driver.fc.Delete(flagName)
	return nil
}
