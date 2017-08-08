package main

import (
	"code.laria.me/petrific/backup"
	"code.laria.me/petrific/cache"
	"code.laria.me/petrific/fs"
	"fmt"
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

	if len(args) != 1 {
		usage()
		return 2
	}

	dir_path, err := abspath(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-dir: %s\n", err)
		return 1
	}

	d, err := fs.OpenOSFile(dir_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-dir: %s\n", err)
		return 1
	}

	if d.Type() != fs.FDir {
		fmt.Fprintf(os.Stderr, "write-dir: %s is not a directory\n", dir_path)
		return 1
	}

	id, err := backup.WriteDir(objectstore, dir_path, d, cache.NopCache{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "write-dir: %s\n", err)
		return 1
	}

	fmt.Println(id)
	return 0
}
