package main

import (
	"code.laria.me/petrific/config"
	"code.laria.me/petrific/storage"
	"flag"
	"fmt"
	"os"
)

type subcmd func(args []string) int

var subcmds = map[string]subcmd{
	"write-dir":       notImplementedYet,
	"restore-dir":     notImplementedYet,
	"take-snapshot":   notImplementedYet,
	"create-snapshot": notImplementedYet,
	"list-snapshots":  notImplementedYet,
}

// Global flags
var (
	flagConfPath = flag.String("config", "", "Use this config file instead of the default")
	flagStorage  = flag.String("storage", "", "Operate on this storage instead of the default one")
)

var conf config.Config
var objectstore storage.Storage

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

	remaining := flag.Args()

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

	storageOptions, ok := conf.Storage[storageName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Storage %s not found\n", storageName)
		return false
	}

	var method string
	if err := storageOptions.Get("method", &method); err != nil {
		fmt.Fprintf(os.Stderr, "Failed setting up storage %s: %s\n", storageName, err)
		return false
	}

	st, ok := storage.StorageTypes[method]
	if !ok {
		fmt.Fprintf(os.Stderr, "Failed setting up storage %s: Method %s unknown", storageName, method)
		return false
	}

	s, err := st(conf, storageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed setting up storage %s: %s\n", storageName, err)
		return false
	}

	objectstore = s
	return true
}

func notImplementedYet(_ []string) int {
	fmt.Fprintln(os.Stderr, "Not implemented yet")
	return 1
}
