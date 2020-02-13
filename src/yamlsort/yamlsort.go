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
	"strconv"
	"strings"
	"unicode"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// version string set by ldflags (git describe)
var version string

var yamlsortUsage = `
yaml sorter. read yaml text from stdin or file, output map key sorted text to stdout or file.
`

//---------------------------------------------------------------------
//  stringMacro class
// helm chart macro value
//
type stringMacro struct {
	value string
}

func (c *stringMacro) setString(arg string) {
}
func (c *stringMacro) getString() string {
	return c.value
}

//---------------------------------------------------------------------
//  yamlsortCmd class
//
type yamlsortCmd struct {
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	inputfilename       string
	outputfilename      string
	inputoutputfilename string
	overridefilename    string
	skipkeys            []string
	selectkeys          []string
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
	f.StringArrayVar(&yamlsort.skipkeys, "skip-key", []string{}, "skip key name in marshal output. (can specify multiple values with --skip-key name --skip-key title)")
	f.StringArrayVar(&yamlsort.selectkeys, "select-key", []string{}, "select key name in marshal output. (can specify multiple values with --select-key name --select-key title)")

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
		result, err2 := c.myOverride(data, dataOverride)
		if err2 != nil {
			return err2
		}
		data = result
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
		// fmt.Fprintln(outputWriter, "---")
		// fmt.Fprintf(outputWriter, "%s%s\n", firstlinestr, "# powered by json.MarshalIndent output")
		fmt.Fprintln(outputWriter, string(outputBytes))

	} else {
		// write yamlsort my marshal
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
	err := c.myMershalRecursive(writer, 0, "", false, data)
	return writer.Bytes(), err
}

// return socre of priority key name  , like "name"
func priorIndex(priorkeys []string, s string) int {
	for i, v := range priorkeys {
		if s == v {
			return i
		}
	}
	return 999999
}

// convert string to int slice, number is convert to one int.
func convertStringToUint64Slice(s string) ([]uint64, error) {
	result := []uint64{}
	digitBuf := []rune{}

	for _, r := range s {
		if unicode.IsDigit(r) {
			digitBuf = append(digitBuf, r)
		} else {
			if len(digitBuf) > 0 {
				i, err := strconv.ParseInt(string(digitBuf), 10, 64)
				if err != nil {
					return result, err
				}
				result = append(result, uint64(i))
				digitBuf = []rune{}
			}
			// string character (rune) is may be 32bit value (unicode 16)
			result = append(result, uint64(r)+0x1000000000000000)
		}
	}
	if len(digitBuf) > 0 {
		i, err := strconv.ParseInt(string(digitBuf), 10, 64)
		if err != nil {
			return result, err
		}
		result = append(result, uint64(i))
		digitBuf = []rune{}
	}
	return result, nil
}

// compair string1 string2 , consider prior key name , and string-number-string key
func compairString(s1 string, s2 string) bool {
	// priority key name check
	score1 := priorIndex(globalpriorkeys, s1)
	score2 := priorIndex(globalpriorkeys, s2)
	if score1 != score2 {
		return score1 < score2
	}

	uint64slice1, err1 := convertStringToUint64Slice(s1)
	uint64slice2, err2 := convertStringToUint64Slice(s2)
	if err1 != nil || err2 != nil {
		return s1 < s2
	}

	// string compair with string-number-string
	len1 := len(uint64slice1)
	len2 := len(uint64slice2)
	for i := 0; i < len1 && i < len2; i++ {
		if uint64slice1[i] != uint64slice2[i] {
			return uint64slice1[i] < uint64slice2[i]
		}
	}
	return len1 < len2
}

func (c *yamlsortCmd) escapeString(value string) string {
	blnDoQuote := false
	blnDoDoubleQuote := false

	// if always quote flag, then quote.
	if c.blnQuoteString {
		blnDoQuote = true
	}

	// if string like boolean , then quote.
	boolArray := [...]string{"true", "false", "yes", "no", "on", "off"}
	for _, s := range boolArray {
		if strings.EqualFold(value, s) {
			blnDoQuote = true
		}
	}

	// if string starts with 0-9 , . , then quote.
	numberArray := [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ",", "!", "@", "#", "%", "&", "*", "|", "`", "[", "]", "{", "}"}
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
	}
	// quote '
	// quote ' .  in quote ' ,  ' is ''
	result := "'" + strings.Replace(value, "'", "''", -1) + "'"
	return result
}

func (c *yamlsortCmd) calcPathMap(path string, key string) string {
	if len(path) == 0 {
		return key
	}
	return path + "." + key
}

func (c *yamlsortCmd) calcPathSlice(path string, index int) string {
	if len(path) == 0 {
		return "[" + strconv.Itoa(index) + "]"
	}
	return path + "[" + strconv.Itoa(index) + "]"
}

func (c *yamlsortCmd) calcPathSliceMap(path string, key string, value string) string {
	if len(path) == 0 {
		return "[" + key + "=" + value + "]"
	}
	return path + "[" + key + "=" + value + "]"
}

func (c *yamlsortCmd) checkSkipKey(path string) bool {
	for _, s := range c.skipkeys {
		if len(s) > 0 {
			if s == path {
				return true
			}
		}
	}
	return false
}

func (c *yamlsortCmd) checkSelectKey(path string) bool {

	// 指定が一つもない場合は常に選択OK
	if len(c.selectkeys) == 0 {
		return true
	}

	// 指定がある場合は、指定されたパスの下だけOK
	for _, s := range c.selectkeys {
		if len(s) > 0 {
			// 正解に続く道ならとりあえず許可する。ここで探索を打ち切ると正解にたどり着けないので。
			if strings.HasPrefix(s, path) {
				return true
			}
			// 正解の下は許可する
			if strings.HasPrefix(path, s) {
				return true
			}
		}
	}

	return false
}

func (c *yamlsortCmd) myMershalRecursive(writer io.Writer, level int, path string, blnParentSlide bool, data interface{}) error {
	if data == nil {
		fmt.Fprintln(writer, "null")
		return nil
	}
	if m, ok := data.(map[string]interface{}); ok {
		// data is map

		// if map has no key , then output {}
		if len(m) == 0 {
			indentstr := c.indentstr(level)
			fmt.Fprintf(writer, "%s%s\n", indentstr, "{}")
			return nil
		}

		// get key list
		var keylist []string
		for k := range m {
			keylist = append(keylist, k)
		}

		// sort map key, but key priorkeys is first
		sort.Slice(keylist, func(idx1, idx2 int) bool {
			return compairString(keylist[idx1], keylist[idx2])
		})

		// recursive call
		for i, k := range keylist {
			v := m[k]
			indentstr := c.indentstr(level)
			// when parent element is slice and print first key value, no need to indent
			if blnParentSlide && i == 0 {
				indentstr = ""
			}
			childpath := c.calcPathMap(path, k)
			// check skip key
			if c.checkSkipKey(childpath) == true {
				continue
			}
			// heck select key
			if c.checkSelectKey(childpath) != true {
				continue
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
			err := c.myMershalRecursive(writer, level+2, childpath, false, v)
			if err != nil {
				return err
			}
		}
		return nil
	} else if a, ok := data.([]interface{}); ok {
		// data is slice

		// if array has no data, then output []
		if len(a) == 0 {
			indentstr := c.indentstr(level)
			fmt.Fprintf(writer, "%s%s\n", indentstr, "[]")
			return nil
		}

		for i, v := range a {
			levelOffset := 0
			if c.blnArrayIndentPlus2 {
				levelOffset = 2
			}
			childpath := c.calcPathSlice(path, i)
			if tmpmap, ok2 := v.(map[string]interface{}); ok2 {
				if tmpname, ok3 := tmpmap["name"]; ok3 {
					if tmpnamestr, ok4 := tmpname.(string); ok4 {
						// sliceの中は name要素を持つmapの場合、特別なpath [name=value]を生成
						childpath = c.calcPathSliceMap(path, "name", tmpnamestr)
					}
				}
			}
			// check skip key
			if c.checkSkipKey(childpath) == true {
				continue
			}
			// heck select key
			if c.checkSelectKey(childpath) != true {
				continue
			}
			fmt.Fprintf(writer, "%s- ", c.indentstr(level-2+levelOffset))
			err := c.myMershalRecursive(writer, level+levelOffset, childpath, true, v)
			if err != nil {
				return err
			}
		}
		return nil
	} else if s, ok := data.(stringMacro); ok {
		// data is stringMacro
		fmt.Fprintln(writer, s.getString())
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

func (c *yamlsortCmd) myOverride(data interface{}, dataOverride interface{}) (interface{}, error) {
	result, err := c.myOverrideRecursive(data, dataOverride)
	return result, err
}

func (c *yamlsortCmd) myOverrideRecursive(data interface{}, dataOverride interface{}) (interface{}, error) {
	if dataOverride == nil {
		return data, nil
	}
	if data == nil {
		data = dataOverride
		return data, nil
	}

	{
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
				return compairString(keylist[idx1], keylist[idx2])
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
				} else if _, ok := v.([]interface{}); ok {
					// value is slice
					//if adest, ok2 := vdest.([]interface{}); ok2 {
					//	// dest is slice, so append slice
					//	adest = append(adest, a...)
					//	// override map
					//	mdest[k] = adest
					//}
				} else {
					// value is normal string/float64/int
					mdest[k] = v
					continue
				}
				result, err := c.myOverrideRecursive(vdest, v)
				if err != nil {
					return data, err
				}
				mdest[k] = result
			}
			return data, nil
		}
	}
	{
		// slice check ( slice - map type )
		adest, ok1 := data.([]interface{})
		a, ok2 := dataOverride.([]interface{})
		if ok1 && ok2 {
			blnOverride := false

			// check slice - map["name"] type
			for _, elem := range a {
				if m, ok3 := elem.(map[string]interface{}); ok3 {
					// slice - map
					name := m["name"]
					for idest, destelem := range adest {
						if mdest, ok4 := destelem.(map[string]interface{}); ok4 {
							// slice - map
							namedest := mdest["name"]
							if _, ok5 := namedest.(string); ok5 {
								if name == namedest {
									result, err := c.myOverrideRecursive(mdest, m)
									if err != nil {
										return data, err
									}
									adest[idest] = result
									blnOverride = true
								}
							}
						}
					}
					if blnOverride == false {
						// append
						adest = append(adest, m)
						blnOverride = true
					}
				} else if s, ok4 := elem.(string); ok4 {
					// check []string
					adest = append(adest, s)
					blnOverride = true
				} else if i, ok4 := elem.(int); ok4 {
					// check []string
					adest = append(adest, i)
					blnOverride = true
				} else if f, ok4 := elem.(float64); ok4 {
					// check []string
					adest = append(adest, f)
					blnOverride = true
				} else if b, ok4 := elem.(bool); ok4 {
					// check []string
					adest = append(adest, b)
					blnOverride = true
				}
			}

			if blnOverride == false {
				fmt.Printf("unknown slice type:%v  data:%v", reflect.TypeOf(data), data)
			}
			return adest, nil
		}
	}
	{
		// slice check ( slice - string/int/float64/bool type )
		adest, ok1 := data.([]string)
		a, ok2 := dataOverride.([]string)
		if ok1 && ok2 {
			for _, k := range a {
				adest = append(adest, k)
				fmt.Println("append []string ", k)
			}
			return adest, nil
		}
	}
	{
		// slice check ( slice - string/int/float64/bool type )
		adest, ok1 := data.([]int)
		a, ok2 := dataOverride.([]int)
		if ok1 && ok2 {
			for _, k := range a {
				adest = append(adest, k)
			}
			return adest, nil
		}
	}
	{
		// slice check ( slice - string/int/float64/bool type )
		adest, ok1 := data.([]float64)
		a, ok2 := dataOverride.([]float64)
		if ok1 && ok2 {
			for _, k := range a {
				adest = append(adest, k)
			}
			return adest, nil
		}
	}
	{
		// slice check ( slice - string/int/float64/bool type )
		adest, ok1 := data.([]bool)
		a, ok2 := dataOverride.([]bool)
		if ok1 && ok2 {
			for _, k := range a {
				adest = append(adest, k)
			}
			return adest, nil
		}
	}

	return data, fmt.Errorf("unknown type:%v  data:%v", reflect.TypeOf(data), data)
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
