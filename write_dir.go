package main

import (
	"code.laria.me/petrific/backup"
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"fmt"
	"os"
	"path"
)

func WriteDir(args []string) int {
	if len(args) != 1 || len(args[0]) == 0 {
		subcmdUsage("write-dir", "directory", nil)()
		return 2
	}

	dir_path := args[0]
	// Make path absolute
	if dir_path[0] != '/' {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "write-dir: %s", err)
			return 1
		}
		dir_path = pwd + "/" + dir_path
	}
	dir_path = path.Clean(dir_path)

	d, err := fs.OpenOSFile(dir_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-dir: %s", err)
		return 1
	}

	if d.Type() != fs.FDir {
		fmt.Fprintf(os.Stderr, "write-dir: %s is not a directory", dir_path)
		return 1
	}

	id, err := backup.WriteDir(objectstore, dir_path, d, cache.NopCache{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-dir: %s", err)
		return 1
	}

	fmt.Println(id)
	return 0
}
