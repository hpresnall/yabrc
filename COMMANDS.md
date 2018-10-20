# Command Reference
In general, every command takes at least one argument: the properties file that defines the configuration of the index.

All commands support the following flags:
* `--debug`: enable verbose logging
* `--ext`: the extension of the index file. This is used to specify the index to operate on. The full index path will be `<savePath>/<baseName>_<ext>`. Defaults to `_current`.

yabrc returns `1` if there were any errors processing the command. Otherwise, it returns `0`.

## Update
Update scans the file system to create or update indexes. If the config file defines an index that does not exist it will create it. By default, this command prompts before moving the existing index and inwriting the new one.
* `-a`, `--autosave`: save the updated index(es) without user confirmation
* `-o`, `--overwrite`: does not move the existing index. The new index is written in place and the old one is _deleted_.
* `-f`, `--fast`: only hash new or updated files. Note that this relaxes the integrity guarantee and will miss bit rot on files that have not changed.
* `--old_ext`: use the given extension as the old index instead of the default `_<YYYYmmDD_HHMMSS>`. Note that this _does not_ delete any existing index that already has this extension.
 
## Compare
Compare checks for differences between two existing indexes. Takes one or two config files as arguments. Returns `1` if there are any differences.
* `--ext2`: the extension of the second index to compare. Defaults to `_current`.

To compare two versions of the same index, specify a single config file and `--ext`, `--ext2` or both.

To compare indexes from two different configurations, specify two config files. Without any extension flags, the `_current` versions will be compared.

## Print
Prints out information about an index.

By default, this prints out basic information about the index but no information about the files in the index.
* `--entries`: print out information about each index file entry.
* `--json`: print out the information about the index and all file entries as JSON.

## Version
Prints out version information.

