package template

import (
	"fmt"
	"reflect"
	"regexp"
	"text/template"
)

var (
	startPos = 1
	endPos   = 1

	special       = regexp.MustCompile("[^a-zA-Z0-9]+")
	templateFuncs = template.FuncMap{
		"goType":       goType,
		"picTag":       picTag,
		"sanitiseName": sanitiseName,
		"indexComment": indexComment,
	}
)

var CopyBook = template.Must(
	template.New("struct").
		Funcs(templateFuncs).
		Parse(`
////////////////////////////////
//     AUTOGENERATED FILE     //
// File generated with go-pic //
////////////////////////////////

// nolint
package tempcopybook

// Copybook{{.Name}} contains a representation of your provided Copybook
type Copybook{{.Name}} struct {
	{{- range $element := .Records}}
		{{sanitiseName $element.Name}} {{goType $element.Picture $element.Occurs}} {{picTag $element.Length $element.Occurs}}{{indexComment $element.Length $element.Occurs}} 
	{{- end}}
}
`))

// goType translates a type into a go type
func goType(t reflect.Kind, i int) string {
	switch t {
	case reflect.String:
		if i > 0 {
			return "[]string"
		}
		return "string"
	case reflect.Int:
		if i > 0 {
			return "[]int"
		}
		return "int"
	default:
		panic(fmt.Sprintf("unrecognized type %v", t))
	}
}

func picTag(l int, i int) string {
	if i > 0 {
		return "`" + fmt.Sprintf("pic:\"%d,%d\"", l, i) + "`"
	}
	return "`" + fmt.Sprintf("pic:\"%d\"", l) + "`"
}

func indexComment(l int, i int) string {
	size := l
	if i > 0 {
		size *= i
	}

	s := startPos
	endPos += size
	startPos = endPos
	return fmt.Sprintf(" // start:%d end:%d", s, endPos-1)
}

func sanitiseName(s string) string {
	return special.ReplaceAllString(s, "")
}