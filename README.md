petrific
========

*petrific - Having the quality of petrifying or turning into stone; causing petrifaction*

petrific is a content-addressable backup/data archival software. It can store your data locally or in the cloud, optionally encrypted using GPG.

It stores your files and directories as objects that are addressed with their cryptographic hash. This deduplicates data and guarantees file integrity. The idea is very similar to Git or Venti. In contrast to these systems, petrific uses human readable and extensible file formats for the objects and uses the SHA3-256 hash instead of SHA1.

Installation
------------

You will need the [Go](https://www.golang.org) compiler, [git](https://www.git-scm.org) and [GPG](https://gnupg.org/) installed on your system. After that, installing petrific is as easy as `go get code.laria.me/petrific`.

Configuration & Usage
---------------------

petrific is configured by a [TOML](https://github.com/toml-lang/toml) file, located at `$XDG_CONFIG_HOME/petrific/config.toml`.

Here is a commented example config. For more details, consult the documentation of the `config` and the `storage` package.

	# This config key defines the default storage backend, as defined below
	default_storage = "local_encrypted"

	[signing]
	# Use this GPG key to sign snapshots
	key = "0123456789ABCDEF0123456789ABCDEF01234567"

	# The storage.* sections define storage backends.
	# Every section must contain the key `method`, the other keys depend on the selected method.
	# For more details see the documentation for the storage package

	[storage.local]
	method="local"
	path="~/.local/share/petrific"

	# This storage encrypts all objects with GPG before storing them into the
	# storage.local storage, we defined before
	[storage.local_encrypted]
	method="filter"
	base="local"
	encode=["gpg", "--encrypt", "-r", "0123456789ABCDEF0123456789ABCDEF01234567"]
	decode=["gpg", "--decrypt"]
	# using method="filter" you can e.g. also implement compression

You can then use the `petrific` command line tool. Use `petrific -help` for a description of subcommands.

Should I use it?
----------------

If you feel comfortable on the command line, perhaps. While petrific contains tests, it still could have a larger test suite. Also it needs to prove itself in real world scenarios. It's performance is not that great, so expect large backups to take a while.

Currently you can never delete backups (see also "Wish list" below), if you have large, fast-changing data, it is therefore probably not a good choice.

Use your own judgement.

Wish list
---------

In no particular order:

* Deletion of backups (needs a garbage collection algorithm to detect unused objects).
* Content-aware "blockification" of files. Right now, a file is simply split into 16MB large blocks. A content-aware splitting process could drastically reduce memory usage.
* More tests.
* Progress indicator of some sorts.
* Mounting snapshots (read-only) via FUSE or 9P.
* Do encryption / signing ourselves instead of firing up a GPG process every time. This should improve performance.

Contributing
------------

You can either send your patches (use [`git format-patch`](https://git-scm.com/docs/git-format-patch)) via email or send a pull request on the [GitHub repo](https://www.github.com/silvasur/petrific).

This software is released under the terms of the WTFPL (see LICENSE file), so you probably should agree with that license when sending patches.
