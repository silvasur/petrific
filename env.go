package main

import (
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/logging"
	"code.laria.me/petrific/storage"
	"code.laria.me/petrific/storage/registry"
	"fmt"
)

// Env provides commonly used objects for subcommands
type Env struct {
	Conf    config.Config
	Store   storage.Storage
	IdCache cache.Cache
	Log     *logging.Log
}

func (e *Env) Close() {
	if e.IdCache != nil {
		e.IdCache.Close()
	}

	if e.Store != nil {
		e.Store.Close()
	}
}

func (env *Env) loadCache() error {
	path := env.Conf.CachePath
	if path == "" {
		return nil
	}

	file_cache := cache.NewFileCache(config.ExpandTilde(path))
	if err := file_cache.Load(); err != nil {
		return fmt.Errorf("Loading cache %s: %s", path, err)
	}

	env.IdCache = file_cache
	return nil
}

func NewEnv(log *logging.Log, confPath, storageName string) (*Env, error) {
	env := new(Env)
	env.Log = log

	var err error

	// Load config

	env.Conf, err = config.LoadConfig(confPath)
	if err != nil {
		return nil, err
	}

	// load storage

	if storageName == "" {
		storageName = env.Conf.DefaultStorage
	}

	env.Store, err = registry.LoadStorage(env.Conf, storageName)
	if err != nil {
		return nil, err
	}

	// Load cache

	if err = env.loadCache(); err != nil {
		env.Close()
		return nil, err
	}

	return env, nil
}
