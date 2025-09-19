package sqlt

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
)

const testArgFormatterTemplate = `
SELECT
  foo, bar, baz
FROM table
WHERE foo = {{ .foo | _sqlt_escapeSql }}
  AND bar = {{ .bar | _sqlt_escapeSql }}
  OR baz = {{ .foo | _sqlt_escapeSql }}
`

const testArgFormatterWant = `
SELECT
  foo, bar, baz
FROM table
WHERE foo = @arg0
  AND bar = @arg1
  OR baz = @arg0
`

func TestArgFormatter(t *testing.T) {
	fmter := &argFormatter{}
	tmpl := template.Must(template.New("root").Funcs(template.FuncMap{
		escapeFuncName: fmter.Format,
	}).Parse(testArgFormatterTemplate))

	b := new(bytes.Buffer)
	err := tmpl.Execute(b, map[string]any{
		"foo": 123,
		"bar": "a string",
	})
	if err != nil {
		t.Errorf("Error executing template: %v.", err)
	}

	if !cmp.Equal(testArgFormatterWant, b.String()) {
		t.Errorf(
			"Bad template result: %s",
			cmp.Diff(testArgFormatterWant, b.String()),
		)
	}

	want := []any{123, "a string"}
	if !cmp.Equal(want, fmter.args) {
		t.Errorf(
			"Bad template args: %s",
			cmp.Diff(want, fmter.args),
		)
	}

	fmter2 := &argFormatter{}
	tpl, _ := tmpl.Clone()
	tpl.Funcs(template.FuncMap{
		escapeFuncName: fmter2.Format,
	})

	b.Reset()
	err = tpl.Execute(b, map[string]any{
		"foo": 123,
		"bar": "a string",
	})
	if err != nil {
		t.Errorf("Error executing template: %v.", err)
	}

	if !cmp.Equal(testArgFormatterWant, b.String()) {
		t.Errorf(
			"Bad template result: %s",
			cmp.Diff(testArgFormatterWant, b.String()),
		)
	}

	if !cmp.Equal(want, fmter2.args) {
		t.Errorf(
			"Bad template args: %s",
			cmp.Diff(want, fmter2.args),
		)
	}
}
