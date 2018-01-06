package main

import (
	"code.laria.me/petrific/backup"
	"code.laria.me/petrific/objects"
	"flag"
	"os"
)

func Fsck(env *Env, args []string) int {
	flags := flag.NewFlagSet(os.Args[0]+" fsck", flag.ContinueOnError)
	full := flags.Bool("full", false, "also check blobs")

	flags.Usage = subcmdUsage("fsck", "[flags] [object-id]", flags)
	errout := subcmdErrout(env.Log, "fsck")

	if err := flags.Parse(args); err != nil {
		errout(err)
		return 2
	}

	var fsck_id *objects.ObjectId = nil

	if len(flags.Args()) > 1 {
		id, err := objects.ParseObjectId(flags.Args()[0])
		if err != nil {
			env.Log.Error().Printf("Could not parse object id: %s", err)
			return 1
		}

		fsck_id = &id
	}

	env.Log.Debug().Printf("id: %v", fsck_id)

	problems := make(chan backup.FsckProblem)
	var err error
	go func() {
		err = backup.Fsck(env.Store, fsck_id, *full, problems, env.Log)
		close(problems)
	}()

	problems_found := false
	for p := range problems {
		env.Log.Warn().Print(p)
		problems_found = true
	}

	if problems_found {
		env.Log.Error().Print("Problems found. See warnings in the log")
	}

	if err != nil {
		env.Log.Error().Print(err)
	}

	if err != nil || problems_found {
		return 1
	}
	return 0
}
