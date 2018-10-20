# yabrc - yet another bit rot checker
yabrc is a file integrity checker designed to protect personal file backups from bit rot. It is similar to programs like [integrit](https://github.com/integrit/integrit) but is written in Go and should be easier to run on Windows.

yabrc is designed to be used for checking file integrity before and after backups of nearly complete file systems or large directory trees. It can easily be used to track changes of a single file system over time or the differences between two file systems. It also assumes that the file system structure is mostly the same across backup storage systems and backups are mostly simple bulk copies. So, yabrc does not support a sophisticated set of file matching rules.

The yabrc program compiles to a single executable with no dependencies so it is portable and easy to deploy. Configuration is done via a single properties file for each backup and scan results are stored in a single index file. Indexes are created by scanning the file system and hashing each file. Once created, the index can be compared to other indexes using the same executable.

## Build & Install
yabrc uses `make` to configure and build. Go version 1.8 or higher is needed along with standard Unix tools like awk, bash and grep. [Go dep](https://golang.github.io/dep/) is used for managing dependencies.

1. Run `make setup` to install dep and gometalinter and download dependencies.
2. Run `make` to lint, vet, test and build executables for Linux, MacOS and Windows.
3. Copy the executable for your architecture to somewhere in your path (e.g. `cp out/out/yabrc-linux-amd64 $GOPATH/bin/yabrc`)

## Running yabrc
See the [command reference](COMMANDS.md) for more information.

### Initial Configuration
yabrc configuration is stored in properties files. You will need to create one for each file system or set of directories that you want to track. There are 4 properties, 2 of which are required:
* `root`: (_required_): the root file system or directory of this index.
* `baseName`: (_required_): the default name of the index files created for this file system, _without_ extensions.
* `savePath`: the default path for saving indexes. Defaults to the location of the config file.
* `ignoredDirs`: a comma separated list of regular expressions. Any directory that matches one of the regexps will be skipped and no files or subdirectories will be added to the index.

Usually you will create a pair of configuration files for each backup: one for the source and one for the target. In general only the `root` value needs be different.

On Windows, note that all file paths are printed with `/`, not `\`. For ease of use, it is recommended that all Windows paths in the config file use `/`, e.g. `C:/Users/foo/Documents`. Internally, the index also uses `/` for all file paths so that Windows and Unix file systems can be compared with each other.

### First Run
To build the initial index, run `yabrc update <config.properties>`. This will scan the given file system, starting at `root` and may take a while, depending on the total size of all the files. When complete, it will write the index to `<savePath>/<baseName>_current`.

### Subsequent Runs
Either periodically, or after making changes to the file system, the index can be updated by re-running `yabrc update <config.properties>`. After the scan is complete, it will compare the indexes and print the differences to std out. It will also move the older index to `<savePath>/<baseName>_<YYYYmmDD_HHMMSS>` and write the new index to `<savePath>/<baseName>_current`.

To compare two indexes after the fact, you can run something like `yabrc compare --ext _<YYYYmmDD_HHMMSS> <config.properties>`. Note that two indexes are specified: one by `--ext`; the other defaults to `_current`. This will compare the current, latest index against a previous one from the given datetime, identified by extension.

To compare indexes from two different file systems, run something like `yabrc compare <fs1.properties> <fs2.properties>`, where two configurations are specified. This will compare the two `<baseName>_current` index files.

When comparing indexes, the following symbols are used in the output the indicate changes to a file:
* `!`: the file does not exist in one of the indexes.
* `>` or `<`: the file size has changed; the direction indicates in which index file is larger. The file hash has also necessarily changed.
* `#`: the file size has not changed, but the hash is different. _This may indicate corruption._

### Faster Scans
For frequent backups, it may make sense to only scan for files that have changed. To do this, run `yabrc update` with the `--fast` flag. This will examine the timestamp and size of the file. Files will only be hashed if either of those values have changed. If not, the existing file hash will be used.

This will significantly speed up scans for large file systems at the expense of explicitly checking every file for corruption.  If running in fast mode, you _must_ run periodic full scans to validate continued file system integrity of files at rest.

### Possible Backup Workflow
1. Run `yabrc update --fast` on the source file system
2. Validate the files that have changed are expected
3. Backup the files to the target file system
4. Run `yabrc update --fast` on the target file system
5. Validate the same files have changed from the source
6. Run `yabrc compare` on the source and target index; they should be the same

## Integrity & Security
yabrc uses Go's implementation of [SHA256](https://golang.org/pkg/crypto/sha256/) to hash file contents. The entire contents of the file are hashed so index update times scale with the size of the files. File metadata is not hashed, but the last modification time and the size of the file are stored in the index.

yabrc relies on SHA256 being able to produce different hashes for 1 bit changes in a file, which is a safe assumption of the algorithm. Hash differences will indicate changes to a file insofar as Go's implementation is correct.

Corruption or tampering of the Go compiler or of yabrc could potentially allow the same hash for different file content. No attempts are made to ensure the integrity of Go's implementation at build time or yabrc's executable at run time. OS level security of the system used to build yabrc as well as all systems storing and running yabrc is critical. Note that it _is_ possible to run yabrc from a directory that itself is indexed but that [may not be enough](http://wiki.c2.com/?TheKenThompsonHack) prevent malicious tampering.

Further, index files  are not protected from tampering. yabrc does not verify index files other than what is required for valid parsing. It is recommended that the yabrc configuration files and indexes are stored in file system that is indexed. If additional protection is needed it is certainly possible to encrypt or sign the indexes, but that is out of scope for yabrc.

