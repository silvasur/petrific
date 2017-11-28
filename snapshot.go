package main

import (
	"code.laria.me/petrific/backup"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/gpg"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"
)

func createSnapshot(archive, comment string, tree_id objects.ObjectId, nosign bool) (objects.ObjectId, error) {
	snapshot := objects.Snapshot{
		Archive: archive,
		Comment: comment,
		Date:    time.Now(),
		Tree:    tree_id,
		Signed:  !nosign,
	}

	var payload []byte
	if nosign {
		payload = snapshot.Payload()
	} else {
		var err error
		payload, err = snapshot.SignedPayload(conf.GPGSigner())
		if err != nil {
			return objects.ObjectId{}, fmt.Errorf("could not sign: %s", err)
		}
	}

	obj := objects.RawObject{
		Type:    objects.OTSnapshot,
		Payload: payload,
	}

	return storage.SetObject(objectstore, obj)
}

func CreateSnapshot(args []string) int {
	flags := flag.NewFlagSet(os.Args[0]+" create-snapshot", flag.ContinueOnError)
	nosign := flags.Bool("nosign", false, "don't sign the snapshot (not recommended)")
	comment := flags.String("comment", "", "comment for the snapshot")

	flags.Usage = subcmdUsage("create-snapshot", "[flags] archive tree-object", flags)
	errout := subcmdErrout("create-snapshot")

	err := flags.Parse(args)
	if err != nil {
		errout(err)
		return 2
	}

	args = flags.Args()
	if len(args) != 2 {
		flags.Usage()
		return 2
	}

	tree_id, err := objects.ParseObjectId(args[1])
	if err != nil {
		errout(fmt.Errorf("invalid tree id: %s\n", err))
		return 1
	}

	snapshot_id, err := createSnapshot(args[0], *comment, tree_id, *nosign)
	if err != nil {
		errout(err)
		return 1
	}

	fmt.Println(snapshot_id)
	return 0
}

func TakeSnapshot(args []string) int {
	flags := flag.NewFlagSet(os.Args[0]+" take-snapshot", flag.ContinueOnError)
	nosign := flags.Bool("nosign", false, "don't sign the snapshot (not recommended)")
	comment := flags.String("comment", "", "comment for the snapshot")

	flags.Usage = subcmdUsage("take-snapshot", "[flags] archive dir", flags)
	errout := subcmdErrout("take-snapshot")

	if err := flags.Parse(args); err != nil {
		errout(err)
		return 2
	}

	args = flags.Args()
	if len(args) != 2 {
		flags.Usage()
		return 2
	}

	dir_path, err := abspath(args[1])
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

	tree_id, err := backup.WriteDir(objectstore, dir_path, d, id_cache)
	if err != nil {
		errout(err)
		return 1
	}

	snapshot_id, err := createSnapshot(args[0], *comment, tree_id, *nosign)
	if err != nil {
		errout(err)
		fmt.Fprintf(os.Stderr, "You can try again by running `%s create-snapshot -c '%s' '%s' '%s'\n`", os.Args[0], *comment, args[0], tree_id)
		return 1
	}

	fmt.Println(snapshot_id)
	return 0
}

type snapshotWithId struct {
	id       objects.ObjectId
	snapshot objects.Snapshot
}

type sortableSnapshots []snapshotWithId

func (s sortableSnapshots) Len() int           { return len(s) }
func (s sortableSnapshots) Less(i, j int) bool { return s[i].snapshot.Date.After(s[j].snapshot.Date) }
func (s sortableSnapshots) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func ListSnapshots(args []string) int {
	// usage := subcmdUsage("list-snapshots", "[archive]", nil)
	errout := subcmdErrout("list-snapshots")

	filter := func(s objects.Snapshot) bool { return true }
	if len(args) > 0 {
		archive := args[1]
		filter = func(s objects.Snapshot) bool {
			return s.Archive == archive
		}
	}

	objids, err := objectstore.List(objects.OTSnapshot)
	if err != nil {
		errout(err)
		return 1
	}

	snapshots := make(sortableSnapshots, 0)

	failed := false
	for _, objid := range objids {
		_snapshot, err := storage.GetObjectOfType(objectstore, objid, objects.OTSnapshot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: list-snapshots: could not get snapshot %s: %s\n", objid, err)
			failed = true
			continue
		}

		snapshot := *_snapshot.(*objects.Snapshot)

		if !filter(snapshot) {
			continue
		}

		snapshots = append(snapshots, snapshotWithId{objid, snapshot})
	}

	sort.Sort(snapshots)

	for _, snapshot_id := range snapshots {
		fmt.Printf("%s\t%s\t%s\n\t%s\n", snapshot_id.snapshot.Archive, snapshot_id.snapshot.Date, snapshot_id.id, snapshot_id.snapshot.Comment)
	}

	if failed {
		return 1
	}
	return 0
}

func RestoreSnapshot(args []string) int {
	var snapshotId objects.ObjectId
	flags := flag.NewFlagSet("restore-snapshot", flag.ContinueOnError)
	flags.Var(&snapshotId, "id", "Object id of a snapshot")
	archive := flags.String("archive", "", "Get latest snapshot for this archive")

	flags.Usage = subcmdUsage("restore-snapshot", "[flags] directory", flags)
	errout := subcmdErrout("restore-snapshot")

	err := flags.Parse(args)
	if err != nil {
		errout(err)
		return 2
	}

	args = flags.Args()
	if len(args) < 1 {
		flags.Usage()
		return 2
	}

	if !snapshotId.Wellformed() && *archive == "" {
		errout(errors.New("Either -id or -archive must be given"))
		flags.Usage()
		return 2
	}

	dir_path, err := abspath(args[0])
	if err != nil {
		errout(err)
		return 1
	}

	_, err = os.Stat(dir_path)
	if err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dir_path, 0755); err != nil {
			errout(err)
			return 1
		}
	}

	root, err := fs.OpenOSFile(dir_path)
	if err != nil {
		errout(err)
		return 1
	}

	if root.Type() != fs.FDir {
		errout(fmt.Errorf("%s is not a directory\n", dir_path))
		return 1
	}

	var snapshot *objects.Snapshot

	if *archive != "" {
		snapshot, err = storage.FindLatestSnapshot(objectstore, *archive)
		if err != nil {
			errout(err)
			return 1
		}
	} else {
		_snapshot, err := storage.GetObjectOfType(objectstore, snapshotId, objects.OTSnapshot)
		if err != nil {
			errout(err)
			return 1
		}
		snapshot = _snapshot.(*objects.Snapshot)
	}

	if err := snapshot.Verify(gpg.Verifyer{}); err != nil {
		errout(fmt.Errorf("verification failed: %s", err))
		return 1
	}

	if err := backup.RestoreDir(objectstore, snapshot.Tree, root); err != nil {
		errout(err)
		return 1
	}
	return 0
}
