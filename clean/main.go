package clean

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/dancannon/gorethink"
	"github.com/fatih/color"
	"github.com/jackc/pgx"
)

func Error(err error) {
	if err != nil && err != gorethink.ErrEmptyResult && err != pgx.ErrNoRows {
		pc, fn, line, _ := runtime.Caller(1)
		errString := getErrDetails(err, pc, fn, line)
		log.Println(errString)
	}
}

func getErrDetails(err error, pc uintptr, fn string, line int) string {
	if err != nil {
		errString := color.RedString("\n--------------------------\n")
		errString += fmt.Sprintf(color.RedString("[error]:")+"\n\nLocation: %s \nFile: %s \nLine: %d \nError: %v", runtime.FuncForPC(pc).Name(), fn, line, err)
		errString += color.RedString("\n--------------------------\n")
		return errString
	}
	return ""
}

func JSON(toBeJSON interface{}, excludedFields ...string) map[string]interface{} {
	data, err := json.Marshal(toBeJSON)
	if err != nil {
		Error(err)
	}
	var out map[string]interface{}
	err = json.Unmarshal(data, &out)

	if err != nil {
		Error(err)
	}

	fields := []string{}
	fields = append(fields, excludedFields...)

	for _, field := range fields {
		nestedFields := strings.Split(field, ".")
		if len(nestedFields) > 1 {
			var d map[string]interface{}
			var valid bool
			d, valid = out[nestedFields[0]].(map[string]interface{})
			if !valid {
				break
			}
			for i, nestedField := range nestedFields {
				if _, okk := d[nestedField]; okk {
					if i != 0 {
						if i+1 == len(nestedFields) {
							delete(d, nestedField)
						} else {
							d, valid = d[nestedField].(map[string]interface{})
							if !valid {
								break
							}
						}
					}
				}
			}
		} else {
			if _, ok := out[field]; ok {
				delete(out, field)
			}
		}
	}
	return out
}

func inArrayStr(v string, array []string) bool {
	if array != nil && len(array) > 0 {
		for _, vv := range array {
			if v == vv {
				return true
			}
		}
	}
	return false
}

func Map(m map[string]interface{}, includedFields ...string) map[string]interface{} {
	if m != nil {
		if includedFields != nil && len(includedFields) > 0 {
			cleanedMap := map[string]interface{}{}
			for k, v := range m {
				if inArrayStr(k, includedFields) {
					cleanedMap[k] = v
				}
			}
			return cleanedMap
		}
	}
	return nil
}
