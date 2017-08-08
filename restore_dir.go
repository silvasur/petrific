package main

import (
	"code.laria.me/petrific/backup"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"fmt"
	"os"
)

func RestoreDir(args []string) int {
	usage := subcmdUsage("restore-dir", "directory object-id", nil)

	if len(args) != 2 {
		usage()
		return 2
	}

	dir_path, err := abspath(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "restore-dir: %s\n", err)
		return 1
	}

	d, err := fs.OpenOSFile(dir_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "restore-dir: %s\n", err)
		return 1
	}

	if d.Type() != fs.FDir {
		fmt.Fprintf(os.Stderr, "restore-dir: %s is not a directory\n", dir_path)
		return 1
	}

	id, err := objects.ParseObjectId(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "restore-dir: %s\n", err)
		return 1
	}

	if err := backup.RestoreDir(objectstore, id, d); err != nil {
		fmt.Fprintf(os.Stderr, "restore-dir: %s\n", err)
		return 1
	}

	return 0
}
