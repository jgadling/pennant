package main

import (
	"errors"
	"sync"
)

// Maintain a copy of the flags we pulled from the storage driver, with
// pre-parsed policies.
type FlagCache struct {
	mu    sync.RWMutex
	flags map[string]*Flag
}

// Initialize the FlagCache
func NewFlagCache() (*FlagCache, error) {
	fc := FlagCache{}
	fc.flags = make(map[string]*Flag)
	return &fc, nil
}

// Get a flag by name
func (fc *FlagCache) Get(flagname string) (*Flag, error) {
	fc.mu.RLock()
	flag, ok := fc.flags[flagname]
	fc.mu.RUnlock()
	if ok {
		return flag, nil
	}
	return nil, errors.New("flag not found")
}

// Replace an existing flag
func (fc *FlagCache) Upsert(flag *Flag) error {
	fc.mu.Lock()
	fc.flags[flag.Name] = flag
	fc.mu.Unlock()
	return nil
}

// Remove a flag
func (fc *FlagCache) Delete(flagname string) error {
	fc.mu.Lock()
	delete(fc.flags, flagname)
	fc.mu.Unlock()
	return nil
}

// Get all flags
func (fc *FlagCache) List() map[string]*Flag {
	return fc.flags
}
