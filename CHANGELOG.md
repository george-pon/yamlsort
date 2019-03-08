# CHANGELOG.md

### version 0.1.16

* fix: quote string value 'TRUE', 'YES', 'ON' always to avoid treated as boolean value.

### version 0.1.15

* fix : zero length array output.
* add --select-key option, select output key from myMarshal outout.
    example : --select-key spec.template.spec.containers[name=kjwikigdocker-container].env[name=abc] \
              --select-key spec.template.spec.containers[name=kjwikigdocker-container].name
    whense spec.template.spec.containers has array of map , which has key 'name' and value  'kjwikigdocker-container'.
    because output - name: kjwikigdocker-container line , you have to set also --select-key containers[name=kjwikigdocker-container].name .

### version 0.1.14

* add --skip-key option. skip output (remove) key from myMarshal output.
    example : --skip-key spec.template.spec.containers[name=kjwikigdocker-container].env[name=abc]
    whense spec.template.spec.containers has array of map , which has key 'name' and value  'kjwikigdocker-container'.

### version 0.1.13

* fix when map has no key, output {}

### version 0.1.12

* fix when string value start with , output quoted.
* fix key name sort order. key9 < key 10

### version 0.1.11

* fix string value true,false,yes,no,on,off with string quote.

### version 0.1.10

* improve support map["name"] in slice override.

### version 0.1.9

* support map["name"] in slice override.

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

