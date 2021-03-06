package main

import (
	"code.laria.me/petrific/logging"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type subcmd func(env *Env, args []string) int

var subcmds = map[string]subcmd{
	"write-dir":        WriteDir,
	"restore-dir":      RestoreDir,
	"take-snapshot":    TakeSnapshot,
	"create-snapshot":  CreateSnapshot,
	"list-snapshots":   ListSnapshots,
	"restore-snapshot": RestoreSnapshot,
	"fsck":             Fsck,
	"storagecmd":       StorageCmd,
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

func subcmdErrout(log *logging.Log, name string) func(error) {
	return func(err error) {
		log.Error().Printf("%s: %s\n", name, err)
	}
}

// Global flags
var (
	flagConfPath  = flag.String("config", "", "Use this config file instead of the default")
	flagStorage   = flag.String("storage", "", "Operate on this storage instead of the default one")
	flagVerbosity = flag.Int("verbosity", int(logging.LWarn), "Verbosity level (0: quiet, 4: everything)")
)

func main() {
	os.Exit(Main())
}

func Main() int {
	rand.Seed(time.Now().UnixNano())

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global flags] command\n\nAvailable commands:\n", os.Args[0])
		for cmd := range subcmds {
			fmt.Fprintf(os.Stderr, "  %s\n", cmd)
		}
		fmt.Fprintln(os.Stderr, "\nGlobal flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	log, err := createLogger(logging.Level(*flagVerbosity))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	env, err := NewEnv(log, *flagConfPath, *flagStorage)
	if err != nil {
		log.Error().Println(err)
		return 1
	}
	defer env.Close()

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

	return cmd(env, remaining[1:])
}

func createLogger(level logging.Level) (log *logging.Log, err error) {
	if level < logging.LQuiet || level > logging.LDebug {
		level = logging.LDebug
		return nil, fmt.Errorf("verbosity %d is out of range", level)
	}

	return logging.NewLog(os.Stderr, level), nil
}
