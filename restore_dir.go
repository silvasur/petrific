package main

import (
	"code.laria.me/petrific/backup"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"fmt"
)

func RestoreDir(env *Env, args []string) int {
	usage := subcmdUsage("restore-dir", "directory object-id", nil)
	errout := subcmdErrout(env.Log, "restore-dir")

	if len(args) != 2 {
		usage()
		return 2
	}

	dir_path, err := abspath(args[0])
	if err != nil {
		errout(err)
		return 1
	}

	d, err := fs.OpenOSFile(dir_path)
	if err != nil {
		errout(err)
		return 1
	}

	if d.Type() != fs.FDir {
		errout(fmt.Errorf("%s is not a directory", dir_path))
		return 1
	}

	id, err := objects.ParseObjectId(args[1])
	if err != nil {
		errout(err)
		return 1
	}

	if err := backup.RestoreDir(env.Store, id, d); err != nil {
		errout(err)
		return 1
	}

	return 0
}
