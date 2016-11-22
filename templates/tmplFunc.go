package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nehmeroumani/pill.go/clean"
)

// recovery will silently swallow all unexpected panics.
func recovery() {
	recover()
}

var striptagsRegexp = regexp.MustCompile("<[^>]*?>")

// Comma produces a string form of the given number in base 10 with
// commas after every three orders of magnitude.
//
// e.g. Comma(834142) -> 834,142
func Comma(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = 0 - v
	}

	parts := []string{"", "", "", "", "", "", ""}
	j := len(parts) - 1

	for v > 999 {
		parts[j] = strconv.FormatInt(v%1000, 10)
		switch len(parts[j]) {
		case 2:
			parts[j] = "0" + parts[j]
		case 1:
			parts[j] = "00" + parts[j]
		}
		v = v / 1000
		j--
	}
	parts[j] = strconv.Itoa(int(v))
	return sign + strings.Join(parts[j:], ",")
}

// FloatComma produces a string form of the given number in base 10 with
// commas after every three orders of magnitude.
//
// e.g. Comma(834142.32) -> 834,142.32
func FloatComma(v float64) string {
	buf := &bytes.Buffer{}
	if v < 0 {
		buf.Write([]byte{'-'})
		v = 0 - v
	}

	parts := strings.Split(strconv.FormatFloat(v, 'f', 2, 64), ".")
	normalNumber, _ := strconv.Atoi(parts[0])
	buf.WriteString(Comma(int64(normalNumber)))

	if len(parts) > 1 {
		buf.Write([]byte{'.'})
		buf.WriteString(parts[1])
	}
	return buf.String()
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 3, 64)
}

func HasField(data interface{}, fieldName string) bool {
	if data != nil && fieldName != "" {
		r := reflect.ValueOf(data)
		v := reflect.Indirect(r).FieldByName(fieldName)
		if v.IsValid() {
			return true
		}
		return false
	}
	return false
}

func ToJSON(toBeJSON interface{}) string {
	json, err := json.Marshal(toBeJSON)
	if err != nil {
		clean.Error(err)
		return ""
	}
	return string(json[:])
}

func Replace(s1 string, s2 string) string {
	defer recovery()

	return strings.Replace(s2, s1, "", -1)
}
func NewLineToBreak(s string) string {
	defer recovery()

	return strings.Replace(s, "\n", "<br />", -1)
}
func Default(arg interface{}, value interface{}) interface{} {
	defer recovery()

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() == 0 {
			return arg
		}
	case reflect.Bool:
		if !v.Bool() {
			return arg
		}
	default:
		return value
	}

	return value
}

func Length(value interface{}) int {
	defer recovery()

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len()
	case reflect.String:
		return len([]rune(v.String()))
	}

	return 0
}
func Lower(s string) string {
	defer recovery()

	return strings.ToLower(s)
}

func Upper(s string) string {
	defer recovery()

	return strings.ToUpper(s)
}

func TruncateChars(n int, s string) string {
	defer recovery()

	if n < 0 {
		return s
	}

	r := []rune(s)
	rLength := len(r)

	if n >= rLength {
		return s
	}

	if n > 3 && rLength > 3 {
		return string(r[:n-3]) + "..."
	}

	return string(r[:n])
}

func URLEncode(s string) string {
	defer recovery()

	return url.QueryEscape(s)
}
func WordCount(s string) int {
	defer recovery()

	return len(strings.Fields(s))
}

func Divisibleby(arg interface{}, value interface{}) bool {
	defer recovery()

	var v float64
	switch value.(type) {
	case int, int8, int16, int32, int64:
		v = float64(reflect.ValueOf(value).Int())
	case uint, uint8, uint16, uint32, uint64:
		v = float64(reflect.ValueOf(value).Uint())
	case float32, float64:
		v = reflect.ValueOf(value).Float()
	default:
		return false
	}

	var a float64
	switch arg.(type) {
	case int, int8, int16, int32, int64:
		a = float64(reflect.ValueOf(arg).Int())
	case uint, uint8, uint16, uint32, uint64:
		a = float64(reflect.ValueOf(arg).Uint())
	case float32, float64:
		a = reflect.ValueOf(arg).Float()
	default:
		return false
	}

	return math.Mod(v, a) == 0
}
func LengthIs(arg int, value interface{}) bool {
	defer recovery()

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == arg
	case reflect.String:
		return len([]rune(v.String())) == arg
	}

	return false
}

func Trim(s string) string {
	defer recovery()

	return strings.TrimSpace(s)
}

func CapFirst(s string) string {
	defer recovery()

	return strings.ToUpper(string(s[0])) + s[1:]
}
func Pluralize(arg string, value interface{}) string {
	defer recovery()

	flag := false
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flag = v.Int() == 1
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		flag = v.Uint() == 1
	default:
		return ""
	}

	if !strings.Contains(arg, ",") {
		arg = "," + arg
	}

	bits := strings.Split(arg, ",")

	if len(bits) > 2 {
		return ""
	}

	if flag {
		return bits[0]
	}

	return bits[1]
}
func YesNo(yes string, no string, value bool) string {
	defer recovery()

	if value {
		return yes
	}

	return no
}
func RJust(arg int, value string) string {
	defer recovery()

	n := arg - len([]rune(value))

	if n > 0 {
		value = strings.Repeat(" ", n) + value
	}

	return value
}
func LJust(arg int, value string) string {
	defer recovery()

	n := arg - len([]rune(value))

	if n > 0 {
		value = value + strings.Repeat(" ", n)
	}

	return value
}
func Center(arg int, value string) string {
	defer recovery()

	n := arg - len([]rune(value))

	if n > 0 {
		left := n / 2
		right := n - left
		value = strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
	}

	return value
}

func FileSizeFormat(value interface{}) string {
	defer recovery()

	var size float64

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		size = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		size = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		size = v.Float()
	default:
		return ""
	}

	var KB float64 = 1 << 10
	var MB float64 = 1 << 20
	var GB float64 = 1 << 30
	var TB float64 = 1 << 40
	var PB float64 = 1 << 50

	filesizeFormat := func(filesize float64, suffix string) string {
		return strings.Replace(fmt.Sprintf("%.1f %s", filesize, suffix), ".0", "", -1)
	}

	var result string
	if size < KB {
		result = filesizeFormat(size, "bytes")
	} else if size < MB {
		result = filesizeFormat(size/KB, "KB")
	} else if size < GB {
		result = filesizeFormat(size/MB, "MB")
	} else if size < TB {
		result = filesizeFormat(size/GB, "GB")
	} else if size < PB {
		result = filesizeFormat(size/TB, "TB")
	} else {
		result = filesizeFormat(size/PB, "PB")
	}

	return result
}

func AlphaNumber(value interface{}) interface{} {
	defer recovery()

	name := [10]string{"one", "two", "three", "four", "five",
		"six", "seven", "eight", "nine"}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < 10 {
			return name[v.Int()-1]
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() < 10 {
			return name[v.Uint()-1]
		}
	}

	return value
}
func IntComma(value interface{}) string {
	defer recovery()

	v := reflect.ValueOf(value)

	var x uint
	minus := false
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < 0 {
			minus = true
			x = uint(-v.Int())
		} else {
			x = uint(v.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x = uint(v.Uint())
	default:
		return ""
	}

	var result string
	for x >= 1000 {
		result = fmt.Sprintf(",%03d%s", x%1000, result)
		x /= 1000
	}
	result = fmt.Sprintf("%d%s", x, result)

	if minus {
		result = "-" + result
	}

	return result
}

func Ordinal(value interface{}) string {
	defer recovery()

	v := reflect.ValueOf(value)

	var x uint
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < 0 {
			return ""
		}
		x = uint(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x = uint(v.Uint())
	default:
		return ""
	}

	suffixes := [10]string{"th", "st", "nd", "rd", "th", "th", "th", "th", "th", "th"}

	switch x % 100 {
	case 11, 12, 13:
		return fmt.Sprintf("%d%s", x, suffixes[0])
	}

	return fmt.Sprintf("%d%s", x, suffixes[x%10])
}

func First(value interface{}) interface{} {
	defer recovery()

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		return string([]rune(v.String())[0])
	case reflect.Slice, reflect.Array:
		return v.Index(0).Interface()
	}

	return ""
}

func Last(value interface{}) interface{} {
	defer recovery()

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		str := []rune(v.String())
		return string(str[len(str)-1])
	case reflect.Slice, reflect.Array:
		return v.Index(v.Len() - 1).Interface()
	}

	return ""
}

func Join(arg string, value []string) string {
	defer recovery()

	return strings.Join(value, arg)
}
func Slice(start int, end int, value interface{}) interface{} {
	defer recovery()

	v := reflect.ValueOf(value)

	if start < 0 {
		start = 0
	}

	switch v.Kind() {
	case reflect.String:
		str := []rune(v.String())

		if end > len(str) {
			end = len(str)
		}

		return string(str[start:end])
	case reflect.Slice:
		return v.Slice(start, end).Interface()
	}
	return ""
}

func Random(value interface{}) interface{} {
	defer recovery()

	rand.Seed(time.Now().UTC().UnixNano())

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		str := []rune(v.String())
		return string(str[rand.Intn(len(str))])
	case reflect.Slice, reflect.Array:
		return v.Index(rand.Intn(v.Len())).Interface()
	}

	return ""
}

func StripTags(s string) string {
	return strings.TrimSpace(striptagsRegexp.ReplaceAllString(s, ""))
}

func IsCPPage(pageName string) bool {
	if pageName != "cp-login" && pageName != "" {
		if len(pageName) > 3 {
			if pageName[:3] == "cp-" {
				return true
			}
		}
	}
	return false
}

func IsActive(value1 string, value2 string, oneClass ...bool) string {
	if value1 == value2 {
		if oneClass != nil && len(oneClass) > 0 {
			if oneClass[0] {
				return ` class="active" `
			}
		}
		return " active"
	}
	return ""
}

func InversePosition(lisLength int, position int) int {
	return lisLength - 1 - position
}

func TJoin(s ...string) string {
	// first arg is sep, remaining args are strings to join
	return strings.Join(s[1:], s[0])
}

func Timestamp(t *time.Time) int64 {
	return t.Unix()
}

func URL(url string, query map[string]string) string {
	if query != nil && len(query) > 0 {
		url = strings.TrimSpace(url)
		if url == "" {
			url = "/"
			firstKey := true
			for k, v := range query {
				if firstKey {
					url += "?"
					firstKey = false
				} else {
					url += "&"
				}
				url += k + "=" + v
			}
		} else if strings.HasPrefix(url, "/") {
			_url := "http://www.example.com" + url
			u, err := url.Parse(_url)
			if err == nil {
				firstKey := true
				if u.RawQuery != "" {
					firstKey = false
				}
				for k, v := range query {
					if firstKey {
						firstKey = false
					} else {
						u.RawQuery += "&"
					}
					u.RawQuery += k + "=" + v
				}
				url = stings.Replace(u.String(), "http://www.example.com", -1)
			}
		} else {
			u, err := url.Parse(_url)
			if err == nil {
				firstKey := true
				if u.RawQuery != "" {
					firstKey = false
				}
				for k, v := range query {
					if firstKey {
						firstKey = false
					} else {
						u.RawQuery += "&"
					}
					u.RawQuery += k + "=" + v
				}
				url = u.String()
			}
		}
	}
	return url
}
