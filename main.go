package main

import (
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/storage"
	"code.laria.me/petrific/storage/registry"
	"flag"
	"fmt"
	"os"
)

type subcmd func(args []string) int

var subcmds = map[string]subcmd{
	"write-dir":        WriteDir,
	"restore-dir":      RestoreDir,
	"take-snapshot":    TakeSnapshot,
	"create-snapshot":  CreateSnapshot,
	"list-snapshots":   ListSnapshots,
	"restore-snapshot": RestoreSnapshot,
}

func subcmdUsage(name string, usage string, flags *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s\n", os.Args[0], name, usage)
		if flags != nil {
			fmt.Fprintln(os.Stderr, "\nFlags:")
			flags.PrintDefaults()
		}
	}
}

func subcmdErrout(name string) func(error) {
	return func(err error) {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
	}
}

// Global flags
var (
	flagConfPath = flag.String("config", "", "Use this config file instead of the default")
	flagStorage  = flag.String("storage", "", "Operate on this storage instead of the default one")
)

var conf config.Config
var objectstore storage.Storage
var id_cache cache.Cache = cache.NopCache{}

func main() {
	os.Exit(Main())
}

func Main() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global flags] command\n\nAvailable commands:\n", os.Args[0])
		for cmd := range subcmds {
			fmt.Fprintf(os.Stderr, "  %s\n", cmd)
		}
		fmt.Fprintln(os.Stderr, "\nGlobal flags:")
		flag.PrintDefaults()
	}
	flag.Parse()
	if !loadConfig() {
		return 1
	}
	defer objectstore.Close()

	if !loadCache(conf) {
		return 1
	}
	defer id_cache.Close()

	remaining := make([]string, 0)
	for _, arg := range flag.Args() {
		if arg != "" {
			remaining = append(remaining, arg)
		}
	}

	var cmd subcmd
	if len(remaining) > 0 {
		cmd = subcmds[remaining[0]]
	}

	if cmd == nil {
		flag.Usage()
		return 1
	}

	return cmd(remaining[1:])
}

func loadCache(conf config.Config) bool {
	if conf.CachePath == "" {
		return true
	}

	file_cache := cache.NewFileCache(config.ExpandTilde(conf.CachePath))
	if err := file_cache.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Loading cache %s: %s", conf.CachePath, err)
		return false
	}

	id_cache = file_cache

	return true
}

func loadConfig() bool {
	var err error
	conf, err = config.LoadConfig(*flagConfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading config: %s\n", err)
		return false
	}

	storageName := *flagStorage
	if storageName == "" {
		storageName = conf.DefaultStorage
	}

	s, err := registry.LoadStorage(conf, storageName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	objectstore = s
	return true
}
