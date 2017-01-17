package templates

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nehmeroumani/pill.go/clean"
)

var tmplFuncs template.FuncMap
var Templates *template.Template

var templatesPath string
var tmplDelims []string

func Setup(tmplsPath string, delims ...string) {
	templatesPath = filepath.FromSlash(tmplsPath)
	if delims != nil && len(delims) > 1 {
		tmplDelims = []string{delims[0], delims[1]}
	}
	GetTemplates()
}

func compileTemplates(filePaths []string) (*template.Template, error) {
	tmpl := template.New("templates")
	if tmplDelims != nil && len(tmplDelims) > 1 {
		tmpl = tmpl.Delims(tmplDelims[0], tmplDelims[1])
	}
	tmpl = tmpl.Funcs(tmplFuncs)
	for _, filePath := range filePaths {
		name := filepath.Base(filePath)
		tmpl = tmpl.New(name)

		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		tmpl.Parse(string(b))
	}
	return tmpl, nil
}

//It Read templates files and return it as a value of type Template
func initializeTemplates() *template.Template {
	var err error

	templateFolder, _ := os.Open(templatesPath)
	defer templateFolder.Close()

	templatesPaths := []string{}
	templatesPathsRaw, _ := templateFolder.Readdir(-1)

	for _, file := range templatesPathsRaw {
		if !file.IsDir() {
			templatesPaths = append(templatesPaths, filepath.Join(templatesPath, file.Name()))
		}
	}
	var tmpl *template.Template
	tmpl, err = compileTemplates(templatesPaths)
	if err != nil {
		clean.Error(err)
	}
	return tmpl
}

func AddTmplFunc(funcName string, function interface{}) {
	if tmplFuncs == nil {
		tmplFuncs = template.FuncMap{funcName: function}
	} else {
		tmplFuncs[funcName] = function
	}
}

func GetTemplates() *template.Template {
	if Templates == nil {
		RegisterTmplFunc()
		Templates = initializeTemplates()
	}
	return Templates
}

func RegisterTmplFunc() {
	AddTmplFunc("FloatComma", FloatComma)
	AddTmplFunc("ToJSON", ToJSON)
	AddTmplFunc("FloatToString", FloatToString)
	AddTmplFunc("Replace", Replace)
	AddTmplFunc("NewLineToBreak", NewLineToBreak)
	AddTmplFunc("Default", Default)
	AddTmplFunc("Length", Length)
	AddTmplFunc("Lower", Lower)
	AddTmplFunc("Upper", Upper)
	AddTmplFunc("TruncateChars", TruncateChars)
	AddTmplFunc("URLEncode", URLEncode)
	AddTmplFunc("WordCount", WordCount)
	AddTmplFunc("Divisibleby", Divisibleby)
	AddTmplFunc("LengthIs", LengthIs)
	AddTmplFunc("Trim", Trim)
	AddTmplFunc("CapFirst", CapFirst)
	AddTmplFunc("Pluralize", Pluralize)
	AddTmplFunc("YesNo", YesNo)
	AddTmplFunc("RJust", RJust)
	AddTmplFunc("LJust", LJust)
	AddTmplFunc("Center", Center)
	AddTmplFunc("FileSizeFormat", FileSizeFormat)
	AddTmplFunc("AlphaNumber", AlphaNumber)
	AddTmplFunc("IntComma", IntComma)
	AddTmplFunc("Ordinal", Ordinal)
	AddTmplFunc("First", First)
	AddTmplFunc("Last", Last)
	AddTmplFunc("URL", URL)
	AddTmplFunc("Join", Join)
	AddTmplFunc("Slice", Slice)
	AddTmplFunc("Random", Random)
	AddTmplFunc("StripTags", StripTags)
	AddTmplFunc("IsCPPage", IsCPPage)
	AddTmplFunc("IsActive", IsActive)
	AddTmplFunc("InversePosition", InversePosition)
	AddTmplFunc("TJoin", TJoin)
	AddTmplFunc("Timestamp", Timestamp)
}
func GetTemplate(templateName string) *template.Template {
	if Templates != nil {
		return Templates.Lookup(templateName)
	}
	return nil
}
