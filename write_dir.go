package main

import (
	"fmt"
	"github.com/silvasur/petrific/backup"
	"github.com/silvasur/petrific/cache"
	"github.com/silvasur/petrific/fs"
	"os"
	"path"
)

func abspath(p string) (string, error) {
	if p[0] != '/' {
		pwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("abspath(%s): %s", p, err)
		}
		p = pwd + "/" + p
	}
	return path.Clean(p), nil
}

func WriteDir(args []string) int {
	usage := subcmdUsage("write-dir", "directory", nil)
	errout := subcmdErrout("write-dir")

	if len(args) != 1 {
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
		errout(fmt.Errorf("%s is not a directory\n", dir_path))
		return 1
	}

	id, err := backup.WriteDir(objectstore, dir_path, d, cache.NopCache{})
	if err != nil {
		errout(err)
		return 1
	}

	fmt.Println(id)
	return 0
}
