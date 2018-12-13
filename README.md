# yamlsort

yamlsort command marshal yaml/json data , with sorting map key name.

usage

```
yamlsort < input.yaml > output.yaml
```

### key name sort output sample

sort map's key name order by a,b,c , but name comes first.

```
cat > sample.yaml << "EOF"
spec:
  ports:
  - b: b-value
    c: c-value
    name: http
    a: a-value
EOF
yamlsort < sample.yaml
```

results

```
---
# myMarshal output
spec:
  ports:
  - name: http
    a: a-value
    b: b-value
    c: c-value
```

### command help

```
$ yamlsort --help

yaml sorter. read yaml text from stdin or file, output map key sorted text to stdout or file.

Usage:
  yamlsort [flags]

Flags:
      --array-indent-plus-2   output array indent + 2 in yaml format
  -h, --help                  help for yamlsort
  -i, --input-file string     path to input file name
      --jsoninput             read JSON data
      --jsonoutput            use json marshal (encoding/json)
      --key stringArray       set prior key name in sort. default prior key is name. (can specify multiple values with --key name --key title)
      --normal                use marshal (github.com/ghodss/yaml)
  -o, --output-file string    path to output file name
      --quote-string          string value is always quoted in output
```

### option

yamlsort has 3 marshal pattern.
1. sorting map key name marshal (default)
2. github.com/ghodss/yaml marshal ( --normal option )
3. encoding/json marshal ( --jsonoutput option )

### build

```
# go 1.11
bash build-cmd.sh modbuild
```
