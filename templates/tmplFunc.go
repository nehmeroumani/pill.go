package templates

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/nehmeroumani/nrgo/clean"
)

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

// Commaf produces a string form of the given number in base 10 with
// commas after every three orders of magnitude.
//
// e.g. Comma(834142.32) -> 834,142.32
func Commaf(v float64) string {
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

func ShortName(name string, maxLength int) string {
	if name != "" {
		nameLength := utf8.RuneCountInString(name)
		if nameLength <= maxLength {
			return name
		}
		name = name[:maxLength]
		space := 0
		for i, char := range name {
			if char == 32 {
				space = i
			}
		}
		if space != 0 {
			name = name[:space]
		}
		return name + "..."
	}
	return ""
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
