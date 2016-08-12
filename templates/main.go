package templates

import (
	"os"
	"text/template"

	"github.com/nehmeroumani/nrgo/clean"
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
	AddTmplFunc("Comma", Comma)
	AddTmplFunc("Commaf", Commaf)
	AddTmplFunc("ToJSON", ToJSON)
	AddTmplFunc("FloatToString", FloatToString)
}
func GetTemplate(templateName string) *template.Template {
	if Templates != nil {
		return Templates.Lookup(templateName)
	}
	return nil
}
