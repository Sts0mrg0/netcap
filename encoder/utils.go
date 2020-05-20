/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package encoder

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/dreadl0ck/netcap/utils"
	"github.com/evilsocket/islazy/tui"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/types"
)

// e.g: 192.168.1.47->165.227.109.154-53032->80
// to 53032->80
func chopTransportIdent(in string) string {
	arr := strings.Split(in, "-")
	if len(arr) != 4 {
		return ""
	}
	return arr[2] + "-" + arr[3]
}

// e.g: 192.168.1.47->165.227.109.154-53032->80
// to 165.227.109.154->192.168.1.47-80->53032
func reverseIdent(i string) string {
	arr := strings.Split(i, "->")
	if len(arr) != 3 {
		return ""
	}
	middle := strings.Split(arr[1], "-")
	if len(middle) != 2 {
		return ""
	}
	return middle[0] + "->" + arr[0] + "-" + arr[2] + "->" + middle[1]
}

// MarkdownOverview dumps a Markdown summary of all available encoders and their fields
func MarkdownOverview() {
	fmt.Println("# NETCAP Overview " + netcap.Version)
	fmt.Println("> Documentation: [docs.netcap.io](https://docs.netcap.io)")
	fmt.Println("## LayerEncoders")

	fmt.Println("|Name|NumFields|Fields|")
	fmt.Println("|----|---------|------|")
	for _, e := range layerEncoderSlice {
		if csv, ok := netcap.InitRecord(e.Type).(types.AuditRecord); ok {
			fmt.Println("|"+pad(e.Layer.String(), 30)+"|", len(csv.CSVHeader()), "|"+strings.Join(csv.CSVHeader(), ", ")+"|")
		}
	}

	fmt.Println("## CustomEncoders")

	fmt.Println("|Name|NumFields|Fields|")
	fmt.Println("|----|---------|------|")
	for _, e := range customEncoderSlice {
		if csv, ok := netcap.InitRecord(e.Type).(types.AuditRecord); ok {
			fmt.Println("|"+pad(e.Name, 30)+"|", len(csv.CSVHeader()), "|"+strings.Join(csv.CSVHeader(), ", ")+"|")
		}
	}
}

func recovery() {
	if r := recover(); r != nil {
		errorsMapMutex.Lock()
		errorsMap[fmt.Sprint(r)]++
		errorsMapMutex.Unlock()
	}
}

func printProgress(current, total int64) {
	if current%5 == 0 {
		clearLine()
		print("flushing http traffic... (" + progress(current, total) + ")")
	}
}

func progress(current, total int64) string {
	percent := (float64(current) / float64(total)) * 100
	return strconv.Itoa(int(percent)) + "%"
}

func clearLine() {
	print("\033[2K\r")
}

func calcMd5(s string) string {

	var out []byte
	for _, b := range md5.Sum([]byte(s)) {
		out = append(out, b)
	}

	return hex.EncodeToString(out)
}

var typeMap = make(map[string]int)
var fieldNameMap = make(map[string]int)

func countFields(t types.Type) int {
	recordFields := 0
	if r, ok := netcap.InitRecord(t).(types.AuditRecord); ok {

		auditRecord := reflect.ValueOf(r).Elem()

		// iterate over audit record fields
		for i := 0; i < auditRecord.NumField(); i++ {

			// get StructField
			field := auditRecord.Type().Field(i)
			fieldNameMap[field.Name]++

			switch field.Type.String() {
			case "string", "int32", "uint32", "bool", "int64", "uint64", "uint8", "float64":
				recordFields++
				//fmt.Println("  ", field.Name, field.Type, "1")
			default:
				if field.Type.Elem().Kind() == reflect.Struct {
					//fmt.Println("  ", field.Name, field.Type, field.Type.Elem().NumField())
					recordFields += field.Type.Elem().NumField()
					typeMap[strings.TrimPrefix(field.Type.String(), "*")] = field.Type.Elem().NumField()
				} else {
					if field.Type.Elem().Kind() == reflect.Ptr {
						recordFields += field.Type.Elem().Elem().NumField()
						//fmt.Println("  ", field.Name, field.Type, field.Type.Elem().Elem().NumField())
						typeMap[strings.TrimPrefix(strings.TrimPrefix(field.Type.String(), "[]"), "*")] = field.Type.Elem().Elem().NumField()
					} else {
						// scalar array types
						//fmt.Println("  ", field.Name, field.Type, "1")
						recordFields++
					}
				}
			}
		}
	}

	typeMap["types."+strings.TrimPrefix(t.String(), "NC_")] = recordFields

	return recordFields
}

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

// Pair describes a key and an associated value
type Pair struct {
	Key   string
	Value int
}

// PairList implements sort.Interface
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func ShowEncoders() {

	var (
		totalFields, totalAuditRecords int
	)

	fmt.Println("custom:", len(customEncoderSlice))
	for _, e := range customEncoderSlice {
		totalAuditRecords++
		f := countFields(e.Type)
		totalFields += f
		fmt.Println("+", e.Type.String(), "(", f, ")")
	}
	fmt.Println("> custom encoder fields: ", totalFields)
	fmt.Println("> custom encoder audit records:", totalAuditRecords)

	fmt.Println("layer:", len(layerEncoderSlice))
	for _, e := range layerEncoderSlice {
		totalAuditRecords++
		f := countFields(e.Type)
		totalFields += f
		fmt.Println("+", e.Layer.String(), "(", f, ")")
	}

	var rows [][]string
	for _, p := range rankByWordCount(typeMap) {
		rows = append(rows, []string{p.Key, strconv.Itoa(p.Value)})
	}
	fmt.Println("\nTypes:")
	tui.Table(os.Stdout, []string{"Type", "NumFields"}, rows)

	rows = [][]string{}
	for _, p := range rankByWordCount(fieldNameMap) {
		rows = append(rows, []string{p.Key, strconv.Itoa(p.Value)})
	}
	fmt.Println("\nFields:")
	tui.Table(os.Stdout, []string{"Name", "Count"}, rows)

	fmt.Println("> total fields: ", totalFields)
	fmt.Println("> total audit records:", totalAuditRecords)
	fmt.Println("> number of unique fields:", len(fieldNameMap))
}

// Entropy returns the shannon entropy value
// https://rosettacode.org/wiki/Entropy#Go
func Entropy(data []byte) (entropy float64) {
	if len(data) == 0 {
		return 0
	}
	for i := 0; i < 256; i++ {
		px := float64(bytes.Count(data, []byte{byte(i)})) / float64(len(data))
		if px > 0 {
			entropy += -px * math.Log2(px)
		}
	}
	return entropy
}

// pad the input up to the given number of space characters
func pad(in interface{}, length int) string {
	return fmt.Sprintf("%-"+strconv.Itoa(length)+"s", in)
}

func logReassemblyError(t string, s string, a ...interface{}) {
	errorsMapMutex.Lock()
	numErrors++
	nb := errorsMap[t]
	errorsMap[t] = nb + 1
	errorsMapMutex.Unlock()

	if c.Debug {
		utils.ReassemblyLog.Printf("ERROR: "+s, a...)
	}
}

func logReassemblyInfo(s string, a ...interface{}) {
	if c.Debug {
		utils.ReassemblyLog.Printf("INFO: "+s, a...)
	}
}

func logReassemblyDebug(s string, a ...interface{}) {
	if c.Debug {
		utils.ReassemblyLog.Printf("DEBUG: "+s, a...)
	}
}
