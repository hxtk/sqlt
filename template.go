package sqlt

import (
	"bytes"
	"errors"
	"io/fs"
	"sync"
	"text/template"

	"github.com/jackc/pgx/v5"
)

var ErrAlreadyUsed = errors.New("cannot parse tempalte after first use")

type Template struct {
	text *template.Template

	escErr  error
	escaped bool
	once    sync.Once
}

func New(name string) *Template {
	return &Template{
		text: template.New(name).Funcs(template.FuncMap{
			escapeFuncName: new(argFormatter).Format,
		}),
	}
}

func (t *Template) Execute(data any) (query string, params pgx.NamedArgs, err error) {
	tpl, fmter, err := t.preExec()
	if err != nil {
		return "", nil, err
	}

	var b bytes.Buffer
	err = tpl.Execute(&b, data)
	if err != nil {
		return "", nil, err
	}

	return b.String(), fmter.Args(), err
}

func (t *Template) ExecuteTemplate(name string, data any) (query string, params pgx.NamedArgs, err error) {
	tpl, fmter, err := t.preExec()
	if err != nil {
		return "", nil, err
	}

	var b bytes.Buffer
	err = tpl.ExecuteTemplate(&b, name, data)
	if err != nil {
		return "", nil, err
	}

	return b.String(), fmter.Args(), err
}

func (t *Template) preExec() (*template.Template, *argFormatter, error) {
	err := t.escape()
	if err != nil {
		return nil, nil, err
	}

	tpl, err := t.Clone()
	if err != nil {
		return nil, nil, err
	}
	var fmter argFormatter
	tpl.Funcs(template.FuncMap{
		escapeFuncName: fmter.Format,
	})

	return tpl.text, &fmter, nil
}

func (t *Template) escape() error {
	t.once.Do(func() {
		if t.escaped {
			return
		}
		if t.text.Tree != nil {
			t.escErr = escapeNode(t.text, t.text.Root)
			if t.escErr != nil {
				return
			}
			for _, v := range t.text.Templates() {
				t.escErr = escapeNode(v, v.Root)
				if t.escErr != nil {
					return
				}
			}
		}
	})
	return t.escErr
}

func (t *Template) Clone() (*Template, error) {
	tpl, err := t.text.Clone()
	if err != nil {
		return nil, err
	}

	return &Template{
		text:    tpl,
		escErr:  t.escErr,
		escaped: t.escaped,
	}, nil
}

func (t *Template) Funcs(funcMap template.FuncMap) *Template {
	t.text.Funcs(funcMap)
	return t
}

func (t *Template) ParseFS(fsys fs.FS, patterns ...string) (*Template, error) {
	if t.escaped {
		return nil, ErrAlreadyUsed
	}

	_, err := t.text.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, err
	}

	return t, nil
}
