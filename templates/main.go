package templates

import (
	"os"
	"text/template"

	"github.com/nehmeroumani/pill.go/clean"
)

var tmplFuncs template.FuncMap
var Templates *template.Template

var templatesPath string

func Setup(tmplsPath string) {
	templatesPath = tmplsPath
	GetTemplates()
}

//It Read templates files and return it as a value of type Template
func initializeTemplates() *template.Template {
	var err error
	tmpl := template.New("templates")

	templateFolder, _ := os.Open(templatesPath)
	defer templateFolder.Close()

	templatesPaths := new([]string)
	templatesPathsRaw, _ := templateFolder.Readdir(-1)

	for _, file := range templatesPathsRaw {
		if !file.IsDir() {
			*templatesPaths = append(*templatesPaths, templatesPath+"/"+file.Name())
		}
	}
	tmpl, err = tmpl.Funcs(tmplFuncs).ParseFiles(*templatesPaths...)
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
	AddTmplFunc("Join", Join)
	AddTmplFunc("Slice", Slice)
	AddTmplFunc("Random", Random)
	AddTmplFunc("StripTags", StripTags)
}
func GetTemplate(templateName string) *template.Template {
	if Templates != nil {
		return Templates.Lookup(templateName)
	}
	return nil
}
