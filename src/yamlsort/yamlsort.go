package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var yamlsortUsage = `
yaml sorter. read yaml text from stdin or file, output map key sorted text to stdout or file.
`

type yamlsortCmd struct {
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	inputfilename       string
	outputfilename      string
	blnInputJSON        bool
	blnNormalMarshal    bool
	blnJSONMarshal      bool
	blnQuoteString      bool
	blnArrayIndentPlus2 bool
	priorkeys           []string
}

func newRootCmd(args []string) *cobra.Command {

	yamlsort := &yamlsortCmd{}

	cmd := &cobra.Command{
		Use:   "yamlsort",
		Short: "yaml sorter",
		Long:  yamlsortUsage,
		RunE: func(c *cobra.Command, args []string) error {
			return yamlsort.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&yamlsort.inputfilename, "input-file", "i", "", "path to input file name")
	f.StringVarP(&yamlsort.outputfilename, "output-file", "o", "", "path to output file name")
	f.BoolVar(&yamlsort.blnInputJSON, "jsoninput", false, "read JSON data")
	f.BoolVar(&yamlsort.blnQuoteString, "quote-string", false, "string value is always quoted in output")
	f.BoolVar(&yamlsort.blnNormalMarshal, "normal", false, "use marshal (github.com/ghodss/yaml)")
	f.BoolVar(&yamlsort.blnJSONMarshal, "jsonoutput", false, "use json marshal (encoding/json)")
	f.BoolVar(&yamlsort.blnArrayIndentPlus2, "array-indent-plus-2", false, "output array indent + 2 in yaml format")
	f.StringArrayVar(&yamlsort.priorkeys, "key", []string{}, "set prior key name in sort. default prior key is name. (can specify multiple values with --key name --key title)")

	yamlsort.stdin = os.Stdin
	yamlsort.stdout = os.Stdout
	yamlsort.stderr = os.Stderr

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// in my marshal, sort prior key
var globalpriorkeys []string

func (c *yamlsortCmd) run() error {

	myReadBytes := []byte{}
	var err error

	// check prior keys
	if len(c.priorkeys) == 0 {
		c.priorkeys = []string{"name"}
	}

	// set global variable priorkeys
	globalpriorkeys = c.priorkeys

	// check input-file option
	if len(c.inputfilename) > 0 {
		// read from file
		myReadBytes, err = ioutil.ReadFile(c.inputfilename)
		if err != nil {
			return err
		}
	} else {
		// read from stdin
		myReadBuffer := new(bytes.Buffer)
		_, err := io.Copy(myReadBuffer, c.stdin)
		if err != nil {
			return err
		}
		myReadBytes = myReadBuffer.Bytes()
	}

	// check output-file option
	outputWriter := c.stdout
	var flushWriter *bufio.Writer
	if len(c.outputfilename) > 0 {
		ofp, err := os.Create(c.outputfilename)
		if err != nil {
			return err
		}
		defer ofp.Close()
		flushWriter = bufio.NewWriter(ofp)
		outputWriter = flushWriter
	}

	// setup file scanner
	reader := bytes.NewReader(myReadBytes)
	scanner := bufio.NewScanner(reader)
	onefilebuffer := new(bytes.Buffer)
	linecount := 0
	firstlinestr := ""
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			linecount = 0

			// flush outfilebuffer
			if onefilebuffer.Len() > 0 {
				// marshal one file
				err = c.procOneFile(outputWriter, firstlinestr, onefilebuffer.Bytes())
				if err != nil {
					return err
				}
				onefilebuffer = new(bytes.Buffer)
				firstlinestr = ""
				if flushWriter != nil {
					err := flushWriter.Flush()
					if err != nil {
						return err
					}
				}
			}
			continue
		}
		linecount++
		if linecount == 1 {
			if len(line) > 0 {
				if strings.HasPrefix(line, "#") {
					firstlinestr = line + "  "
				}
			}
		}
		fmt.Fprintln(onefilebuffer, line)
	}
	// flush outfilebuffer
	if onefilebuffer.Len() > 0 {
		// marshal one file
		err = c.procOneFile(outputWriter, firstlinestr, onefilebuffer.Bytes())
		if err != nil {
			return err
		}
		onefilebuffer = new(bytes.Buffer)
		firstlinestr = ""
		if flushWriter != nil {
			err := flushWriter.Flush()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *yamlsortCmd) procOneFile(outputWriter io.Writer, firstlinestr string, inputbytes []byte) error {
	var data interface{}

	if c.blnInputJSON {
		// parse json data
		err := json.Unmarshal(inputbytes, &data)
		if err != nil {
			fmt.Fprintln(c.stderr, "Unmarshal JSON error:", err)
			return err
		}
	} else {
		// parse yaml data
		err := yaml.Unmarshal(inputbytes, &data)
		if err != nil {
			fmt.Fprintln(c.stderr, "Unmarshal YAML error:", err)
			return err
		}
	}

	if c.blnNormalMarshal {
		// write yaml data with normal marshal (github.com/ghodss/yaml)
		outputBytes, err := yaml.Marshal(data)
		if err != nil {
			fmt.Fprintln(c.stderr, "Marshal error:", err)
			return err
		}
		fmt.Fprintln(outputWriter, "---")
		fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# github.com/ghodss/yaml/Marshal output")
		fmt.Fprintln(outputWriter, string(outputBytes))
	} else if c.blnJSONMarshal {
		// write json data with normal marshal
		outputBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintln(c.stderr, "Marshal error:", err)
			return err
		}
		fmt.Fprintln(outputWriter, "---")
		fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# json.MarshalIndent output")
		fmt.Fprintln(outputWriter, string(outputBytes))

	} else {
		// write yamlsort marshal
		outputBytes2, err := c.myMarshal(data)
		if err != nil {
			fmt.Fprintln(c.stderr, "myMarshal error:", err)
			return err
		}
		fmt.Fprintln(outputWriter, "---")
		fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# myMarshal output")
		fmt.Fprintln(outputWriter, string(outputBytes2))
	}

	return nil
}

func (c *yamlsortCmd) myMarshal(data interface{}) ([]byte, error) {
	// create buffer
	writer := new(bytes.Buffer)
	err := c.myMershalRecursive(writer, 0, false, data)
	return writer.Bytes(), err
}

func priorIndex(priorkeys []string, s string) int {
	for i, v := range priorkeys {
		if s == v {
			return i
		}
	}
	return 999999
}

func (c *yamlsortCmd) myMershalRecursive(writer io.Writer, level int, blnParentSlide bool, data interface{}) error {
	if data == nil {
		fmt.Fprintln(writer, "")
		return nil
	}
	if m, ok := data.(map[string]interface{}); ok {
		// data is map
		// get key list
		var keylist []string
		for k := range m {
			keylist = append(keylist, k)
		}
		// sort map key, but key priorkeys is first
		sort.Slice(keylist, func(idx1, idx2 int) bool {
			score1 := priorIndex(globalpriorkeys, keylist[idx1])
			score2 := priorIndex(globalpriorkeys, keylist[idx2])
			if score1 != score2 {
				return score1 < score2
			}
			return keylist[idx1] < keylist[idx2]
		})
		// recursive call
		for i, k := range keylist {
			v := m[k]
			indentstr := c.indentstr(level)
			// when parent element is slice and print first key value, no need to indent
			if blnParentSlide && i == 0 {
				indentstr = ""
			}
			if v == nil {
				// child is nil. print key only.
				fmt.Fprintf(writer, "%s%s:", indentstr, k)
			} else if _, ok := v.(map[string]interface{}); ok {
				// child is map
				fmt.Fprintf(writer, "%s%s:\n", indentstr, k)
			} else if _, ok := v.([]interface{}); ok {
				// child is slice
				fmt.Fprintf(writer, "%s%s:\n", indentstr, k)
			} else {
				// child is normal string
				fmt.Fprintf(writer, "%s%s: ", indentstr, k)
			}
			err := c.myMershalRecursive(writer, level+2, false, v)
			if err != nil {
				return err
			}
		}
		return nil
	} else if a, ok := data.([]interface{}); ok {
		// data is slice
		for _, v := range a {
			levelOffset := 0
			if c.blnArrayIndentPlus2 {
				levelOffset = 2
			}
			fmt.Fprintf(writer, "%s- ", c.indentstr(level-2+levelOffset))
			err := c.myMershalRecursive(writer, level+levelOffset, true, v)
			if err != nil {
				return err
			}
		}
		return nil
	} else if s, ok := data.(string); ok {
		// data is string
		if c.blnQuoteString {
			// string is always quoted
			fmt.Fprintf(writer, "\"%s\"\n", s)
		} else {
			fmt.Fprintln(writer, s)
		}
	} else if i, ok := data.(int); ok {
		// data is int
		fmt.Fprintln(writer, i)
	} else if f64, ok := data.(float64); ok {
		// data is float64
		fmt.Fprintln(writer, f64)
	} else if b, ok := data.(bool); ok {
		// data is bool
		fmt.Fprintln(writer, b)
	} else {
		return fmt.Errorf("unknown type:%v  data:%v", reflect.TypeOf(data), data)
	}
	return nil
}

func (c *yamlsortCmd) indentstr(level int) string {
	result := ""
	for i := 0; i < level; i++ {
		result = result + " "
	}
	return result
}
