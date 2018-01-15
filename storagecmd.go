package main

import (
	"fmt"
	"os"
)

func StorageCmd(env *Env, args []string) int {
	cmds := env.Store.Subcmds()

	if len(args) == 0 {
		if len(cmds) == 0 {
			fmt.Fprintln(os.Stderr, "No storage subcommands available")
			return 2
		}

		fmt.Fprintln(os.Stderr, "Availabe storage subcommands:")
		for name := range cmds {
			fmt.Fprintln(os.Stderr, "  "+name)
		}
		return 2
	}

	cmd, ok := cmds[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown storage subcommand %s\n", args[0])
	}

	return cmd(args[1:], env.Log, env.Conf)
}
