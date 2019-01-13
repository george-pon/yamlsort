# CHANGELOG.md

### version 0.1.8

* comment output is override.

### version 0.1.7

* add -f option. same as -i and -o.

### version 0.1.6

* fix string escape in YAML format. quoting starts with 0-9 string.

### version 0.1.5

* fix string escape in YAML format

### version 0.1.4

* fix zero length string value output to "".
* yamlsort supports the same input-file output-file. ( buffering in memory, write to file at last. )
* add --version option. displays version number and exit.

### version 0.1.3

* add --override-file option. read from input-file or stdin, merge map by key name from --override-file .

