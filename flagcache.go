package main

import (
	"errors"
	"sync"
)

type FlagCache struct {
	mu    sync.RWMutex
	flags map[string]*Flag
}

func NewFlagCache() (*FlagCache, error) {
	fc := FlagCache{}
	fc.flags = make(map[string]*Flag)
	return &fc, nil
}

func (fc *FlagCache) Get(flagname string) (*Flag, error) {
	fc.mu.RLock()
	flag, ok := fc.flags[flagname]
	fc.mu.RUnlock()
	if ok {
		return flag, nil
	}
	return nil, errors.New("flag not found")
}

func (fc *FlagCache) Upsert(flag *Flag) error {
	fc.mu.Lock()
	fc.flags[flag.Name] = flag
	fc.mu.Unlock()
	return nil
}

func (fc *FlagCache) Delete(flagname string) error {
	fc.mu.Lock()
	delete(fc.flags, flagname)
	fc.mu.Unlock()
	return nil
}

func (fc *FlagCache) List() map[string]*Flag {
	return fc.flags
}
