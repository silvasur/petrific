package backup

import (
	"code.laria.me/petrific/acl"
	"code.laria.me/petrific/fs"
	"code.laria.me/petrific/objects"
	"code.laria.me/petrific/storage"
	"fmt"
	"io"
	"os"
)

func RestoreFile(s storage.Storage, id objects.ObjectId, w io.Writer) error {
	file, err := storage.GetObjectOfType(s, id, objects.OTFile)
	if err != nil {
		return err
	}

	for i, fragment := range *file.(*objects.File) {
		blob_obj, err := storage.GetObjectOfType(s, fragment.Blob, objects.OTBlob)
		if err != nil {
			return err
		}
		blob := *blob_obj.(*objects.Blob)

		if uint64(len(blob)) != fragment.Size {
			return fmt.Errorf("RestoreFile: blob size of %s doesn't match size in fragment %d of file %s", fragment.Blob, i, id)
		}

		if _, err := w.Write(blob); err != nil {
			return err
		}
	}

	return nil
}

func execBitFromACL(a acl.ACL) bool {
	return a.ToUnixPerms()&0100 != 0
}

func RestoreDir(s storage.Storage, id objects.ObjectId, root fs.Dir) error {
	tree_obj, err := storage.GetObjectOfType(s, id, objects.OTTree)
	tree := tree_obj.(objects.Tree)

	seen := make(map[string]struct{})

	for name, file_info := range tree {
		switch file_info.Type() {
		case objects.TETFile:
			tmpname := fmt.Sprintf(".petrific-%d-%s", os.Getpid(), id)
			new_file, err := root.CreateChildFile(tmpname, execBitFromACL(file_info.ACL()))
			if err != nil {
				return err
			}
			rwc, err := new_file.Open()
			if err != nil {
				return err
			}

			if err := RestoreFile(s, file_info.(objects.TreeEntryFile).Ref, rwc); err != nil {
				rwc.Close()
				return err
			}
			rwc.Close()

			if err := root.RenameChild(tmpname, name); err != nil {
				return err
			}
		case objects.TETDir:
			var subdir fs.Dir

			// Try to use existing directory
			child, err := root.GetChild(name)
			if err == nil {
				if child.Type() == fs.FDir {
					subdir = child.(fs.Dir)
				} else {
					if err := child.Delete(); err != nil {
						return err
					}
				}
			} else if err != os.ErrNotExist {
				return err
			}

			// Create directory, if it doesn't exist
			if subdir == nil {
				subdir, err = root.CreateChildDir(name)
				if err != nil {
					return err
				}
			}

			if err := RestoreDir(s, file_info.(objects.TreeEntryDir).Ref, subdir); err != nil {
				return err
			}
		case objects.TETSymlink:
			// Is there already a child of that name? If yes, delete it
			child, err := root.GetChild(name)
			if err == nil {
				if err := child.Delete(); err != nil {
					return err
				}
			} else if err != os.ErrNotExist {
				return err
			}

			if _, err := root.CreateChildSymlink(name, file_info.(objects.TreeEntrySymlink).Target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("child '%s' of %s has unknown tree entry type %s", name, id, file_info.Type())
		}

		seen[name] = struct{}{}
	}

	// We now restored all children, we now need to remove the children of root, that shouldn't be there accoring to the backup
	children, err := root.Readdir()
	if err != nil {
		return err
	}
	for _, c := range children {
		_, ok := seen[c.Name()]
		if !ok {
			if err := c.Delete(); err != nil {
				return err
			}
		}
	}

	return nil
}
