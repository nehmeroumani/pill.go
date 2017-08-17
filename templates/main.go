package templates

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	tmplNameStartIndex := len(strings.TrimPrefix(templatesPath, "./")) + 1
	for _, fp := range filePaths {
		name := fp[tmplNameStartIndex:]
		tmpl = tmpl.New(name)
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			clean.Error(err)
		} else {
			if _, err = tmpl.Parse(string(b)); err != nil {
				clean.Error(err)
			}
		}
	}
	return tmpl, nil
}

//It Read templates files and return it as a value of type Template
func initializeTemplates() *template.Template {
	templatesPaths := []string{}

	err := filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			templatesPaths = append(templatesPaths, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
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
	AddTmplFunc("URLPath", URLPath)
	AddTmplFunc("YoutubeVideoID", YoutubeVideoID)
	AddTmplFunc("IsSelectedNumVal", IsSelectedNumVal)
}
func GetTemplate(templateName string) *template.Template {
	if Templates == nil {
		GetTemplates()
	}
	if Templates != nil {
		return Templates.Lookup(templateName)
	}
	return nil
}
