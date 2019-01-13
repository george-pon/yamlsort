//
// yamlsort - sort by map's key
//
//
//
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

// version string set by ldflags (git describe)
var version string

var yamlsortUsage = `
yaml sorter. read yaml text from stdin or file, output map key sorted text to stdout or file.
`

type yamlsortCmd struct {
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	inputfilename       string
	outputfilename      string
	inputoutputfilename string
	overridefilename    string
	blnInputJSON        bool
	blnNormalMarshal    bool
	blnJSONMarshal      bool
	blnQuoteString      bool
	blnArrayIndentPlus2 bool
	priorkeys           []string
	blnVersion          bool
	version             string
}

func newRootCmd(args []string) *cobra.Command {

	yamlsort := &yamlsortCmd{
		version: version,
	}

	cmd := &cobra.Command{
		Use:   "yamlsort",
		Short: "yaml sorter",
		Long:  yamlsortUsage,
		RunE: func(c *cobra.Command, args []string) error {
			return yamlsort.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&yamlsort.inputoutputfilename, "input-output-file", "f", "", "path to input/output file name")
	f.StringVarP(&yamlsort.inputfilename, "input-file", "i", "", "path to input file name")
	f.StringVarP(&yamlsort.outputfilename, "output-file", "o", "", "path to output file name")
	f.StringVarP(&yamlsort.overridefilename, "override-file", "", "", "path to override input file name")
	f.BoolVar(&yamlsort.blnInputJSON, "jsoninput", false, "read JSON data")
	f.BoolVar(&yamlsort.blnQuoteString, "quote-string", false, "string value is always quoted in output")
	f.BoolVar(&yamlsort.blnNormalMarshal, "normal", false, "use marshal (github.com/ghodss/yaml)")
	f.BoolVar(&yamlsort.blnJSONMarshal, "jsonoutput", false, "use json marshal (encoding/json)")
	f.BoolVar(&yamlsort.blnArrayIndentPlus2, "array-indent-plus-2", false, "output array indent + 2 in yaml format")
	f.BoolVar(&yamlsort.blnVersion, "version", false, "displays version")
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

//------------------------------------------------------------------------
// run main
//
func (c *yamlsortCmd) run(args []string) error {

	if args != nil {
		if len(args) > 0 {
			fmt.Fprintln(c.stdout, "yamlsort version "+c.version)
			return nil
		}
	}
	if c.blnVersion {
		fmt.Fprintln(c.stdout, "yamlsort version "+c.version)
		return nil
	}

	// override inputoutputfilename
	if len(c.inputoutputfilename) > 0 {
		if len(c.inputfilename) == 0 {
			c.inputfilename = c.inputoutputfilename
		}
		if len(c.outputfilename) == 0 {
			c.outputfilename = c.inputoutputfilename
		}
	}

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

	// create output buffer
	outputBuffer := new(bytes.Buffer)

	// setup file scanner
	reader := bytes.NewReader(myReadBytes)
	scanner := bufio.NewScanner(reader)
	onefilebuffer := new(bytes.Buffer)
	linecount := 0
	firstlinestr := ""
	if len(c.inputfilename) > 0 {
		firstlinestr = "# " + c.inputfilename + "  "
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			linecount = 0

			// flush outfilebuffer
			if onefilebuffer.Len() > 0 {
				// marshal one file
				err = c.procOneFile(outputBuffer, firstlinestr, onefilebuffer.Bytes())
				if err != nil {
					return err
				}
				onefilebuffer = new(bytes.Buffer)
				firstlinestr = ""
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
		err = c.procOneFile(outputBuffer, firstlinestr, onefilebuffer.Bytes())
		if err != nil {
			return err
		}
		onefilebuffer = new(bytes.Buffer)
		firstlinestr = ""
	}

	// at last, write outputBuffer into file or stdout.
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
	// do output
	fmt.Fprint(outputWriter, outputBuffer)
	// flush
	if flushWriter != nil {
		err := flushWriter.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

//-------------------------------------------------------------------------------------
//  unmarshal and sort and marshal.
//
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

	// override
	if len(c.overridefilename) > 0 {
		dataOverride, err := c.myLoadFromFile(c.overridefilename)
		if err != nil {
			return err
		}
		if err2 := c.myOverride(data, dataOverride); err2 != nil {
			return err2
		}
	}

	// if firstline contains '# powered by ' , remove it.
	idx := strings.Index(firstlinestr, "# powered by ")
	if idx >= 0 {
		firstlinestr = string([]rune(firstlinestr)[:idx])
	}
	if c.blnNormalMarshal {
		// write yaml data with normal marshal (github.com/ghodss/yaml)
		outputBytes, err := yaml.Marshal(data)
		if err != nil {
			fmt.Fprintln(c.stderr, "Marshal error:", err)
			return err
		}
		fmt.Fprintln(outputWriter, "---")
		fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# powered by github.com/ghodss/yaml/Marshal")
		fmt.Fprintln(outputWriter, string(outputBytes))
	} else if c.blnJSONMarshal {
		// write json data with normal marshal
		outputBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintln(c.stderr, "Marshal error:", err)
			return err
		}
		fmt.Fprintln(outputWriter, "---")
		fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# powered by json.MarshalIndent output")
		fmt.Fprintln(outputWriter, string(outputBytes))

	} else {
		// write yamlsort marshal
		outputBytes2, err := c.myMarshal(data)
		if err != nil {
			fmt.Fprintln(c.stderr, "myMarshal error:", err)
			return err
		}
		fmt.Fprintln(outputWriter, "---")
		fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# powered by myMarshal output")
		fmt.Fprintln(outputWriter, string(outputBytes2))
	}

	return nil
}

//-----------------------------------------------------------------------------------
// my marshal (data to string with sorting map key)
//
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

func (c *yamlsortCmd) escapeString(value string) string {
	blnDoQuote := false
	blnDoDoubleQuote := false

	// if always quote flag, then quote.
	if c.blnQuoteString {
		blnDoQuote = true
	}

	// if string starts with 0-9 , then quote.
	numberArray := [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for _, s := range numberArray {
		if strings.HasPrefix(value, s) {
			blnDoQuote = true
		}
	}

	// if string contains " or ' , then quote.
	if strings.Contains(value, "\"") || strings.Contains(value, "'") {
		blnDoQuote = true
	}

	// if string contains \r \n \t , then quote.
	if strings.Contains(value, "\r") || strings.Contains(value, "\n") || strings.Contains(value, "\t") {
		blnDoQuote = true
		blnDoDoubleQuote = true
	}

	// if string contains { or } , then quote.
	if strings.Contains(value, "{") || strings.Contains(value, "}") {
		blnDoQuote = true
	}

	// if string starts space , then quote.
	if strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		blnDoQuote = true
	}

	// if string starts tab , then quote.
	if strings.HasPrefix(value, "\t") || strings.HasSuffix(value, "\t") {
		blnDoQuote = true
		blnDoDoubleQuote = true
	}

	// if string length == 0 ,  then quote
	if len(value) == 0 {
		blnDoQuote = true
	}
	if !blnDoQuote {
		return value
	}

	if blnDoDoubleQuote {
		// quote "
		result := value
		result = strings.Replace(result, "\\", "\\\\", -1)
		result = strings.Replace(result, "\"", "\\\"", -1)
		result = strings.Replace(result, "\t", "\\t", -1)
		result = strings.Replace(result, "\n", "\\n", -1)
		result = strings.Replace(result, "\r", "\\r", -1)
		result = "\"" + result + "\""
		return result
	} else {
		// quote '
		// quote ' .  in quote ' ,  ' is ''
		result := "'" + strings.Replace(value, "'", "''", -1) + "'"
		return result
	}
}

func (c *yamlsortCmd) myMershalRecursive(writer io.Writer, level int, blnParentSlide bool, data interface{}) error {
	if data == nil {
		fmt.Fprintln(writer, "null")
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
				fmt.Fprintf(writer, "%s%s: ", indentstr, k)
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
		fmt.Fprintln(writer, c.escapeString(s))
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

//-------------------------------------------------------------------------
// my Override
//
func (c *yamlsortCmd) myOverride(data interface{}, dataOverride interface{}) error {
	return c.myOverrideRecursive(data, dataOverride)
}

func (c *yamlsortCmd) myOverrideRecursive(data interface{}, dataOverride interface{}) error {
	if dataOverride == nil {
		return nil
	}
	if data == nil {
		data = dataOverride
		return nil
	}

	// map check
	mdest, ok1 := data.(map[string]interface{})
	m, ok2 := dataOverride.(map[string]interface{})
	if ok1 && ok2 {
		// dataOverride is map
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
		for _, k := range keylist {
			vdest := mdest[k]
			v := m[k]
			// vdest is nil, then copy and continue
			if vdest == nil {
				mdest[k] = v
				continue
			}
			// when parent element is slice and print first key value, no need to indent
			if v == nil {
				// value is nil. key only.
				mdest[k] = v
				continue
			} else if _, ok := v.(map[string]interface{}); ok {
				// value is map
			} else if a, ok := v.([]interface{}); ok {
				// value is slice
				if adest, ok2 := vdest.([]interface{}); ok2 {
					// dest is slice, so append slice
					adest = append(adest, a...)
					// override map
					mdest[k] = adest
				}
			} else {
				// value is normal string/float64/int
				mdest[k] = v
				continue
			}
			err := c.myOverrideRecursive(vdest, v)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

//-------------------------------------------------------------------------
// load yaml data from file
//
func (c *yamlsortCmd) myLoadFromFile(filename string) (interface{}, error) {
	var data interface{}
	// read from file
	myReadBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return data, err
	}

	if c.blnInputJSON {
		// parse json data
		err := json.Unmarshal(myReadBytes, &data)
		if err != nil {
			fmt.Fprintln(c.stderr, "Unmarshal JSON error:", err)
			return data, err
		}
	} else {
		// parse yaml data
		err := yaml.Unmarshal(myReadBytes, &data)
		if err != nil {
			fmt.Fprintln(c.stderr, "Unmarshal YAML error:", err)
			return data, err
		}
	}
	return data, nil
}
