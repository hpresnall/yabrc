# Command Reference
In general, every command takes at least one argument: the properties file that defines the configuration of the index.

All commands support the following flags:
* `--debug`: enable verbose logging
* `--ext`: the extension of the index file. This is used to specify the index to operate on. The full index path will be `<savePath>/<baseName><ext>`. Defaults to `_current`.

yabrc returns `1` if there were any errors processing the command. Otherwise, it returns `0`.

## `yabrc update`
Update scans the file system to create or update indexes. If the config file defines an index that does not exist it will create it. By default, this command prompts before moving the existing index and writing the new one.
* `-a`, `--autosave`: save the index(es) without user confirmation
* `-o`, `--overwrite`: does not move the existing index. The new index is written in place and the old one is _deleted_.
* `-f`, `--fast`: only hash new or updated files. Note that this relaxes the integrity guarantee and will miss bit rot on files which have not changed size or last update time.
* `--old_ext`: use the given extension as the old index instead of the default `_<YYYYmmDD_HHMMSS>`. This has no effect if `--overwrite` is specified.
 
## `yabrc compare`
Compare checks for differences between two existing indexes. Takes one or two config files as arguments. Returns `1` if there are any differences.
* `--ext2`: the extension of the second index to compare. Defaults to `_current`.

To compare two versions of the same index, specify a single config file and `--ext`, `--ext2` or both.

To compare indexes from two different configurations, specify two config files. Without any extension flags, the `_current` versions will be compared.

The following symbols are used in the output the indicate changes to a file:
* `!`: the file does not exist in one of the indexes.
* `>` or `<`: the file size has changed; the direction indicates in which index file is larger. The file hash has also necessarily changed.
* `#`: the file size has not changed, but the hash is different. _This may indicate corruption._

## `yabrc print`
Prints out information about an index.

By default, this prints out basic information about the index but no information about the files in the index.
* `--entries`: print out information about each index file entry.
* `--json`: print out the information about the index and all file entries as JSON.

## `yabrc version`
Prints out version information.
