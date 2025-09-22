package sqlt

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// dummyX binds values to @x
type dummyX struct {
	args pgx.NamedArgs
}

func (d *dummyX) Format(arg any) string {
	if d.args == nil {
		d.args = pgx.NamedArgs{}
	}
	d.args["x"] = arg
	return "@x"
}
func (d *dummyX) Args() pgx.NamedArgs { return d.args }

// dummyY binds values to @y
type dummyY struct {
	args pgx.NamedArgs
}

func (d *dummyY) Format(arg any) string {
	if d.args == nil {
		d.args = pgx.NamedArgs{}
	}
	d.args["y"] = arg
	return "@y"
}

func (d *dummyY) Args() pgx.NamedArgs { return d.args }

func TestExecute_DefaultEscape(t *testing.T) {
	tpl, err := New("q").Parse(`SELECT * FROM users WHERE id={{ .ID }}`)
	require.NoError(t, err)

	query, args, err := tpl.Execute(map[string]any{"ID": 123})
	require.NoError(t, err)

	require.Equal(t, "SELECT * FROM users WHERE id=@sqlt0", query)
	require.Equal(t, pgx.NamedArgs{"sqlt0": 123}, args)
}

func TestCollectArgs_MultipleCustomSanitizers(t *testing.T) {
	tpl := New("q").Sanitizers(SanitizerMap{
		"safeX": func() Sanitizer { return &dummyX{} },
		"safeY": func() Sanitizer { return &dummyY{} },
	})

	_, err := tpl.Parse(`WHERE a={{ .A | safeX }} AND b={{ .B | safeY }}`)
	require.NoError(t, err)

	query, args, err := tpl.Execute(map[string]any{
		"A": "foo",
		"B": "bar",
	})
	require.NoError(t, err)

	require.Equal(t, "WHERE a=@x AND b=@y", query)
	require.Equal(t, pgx.NamedArgs{
		"x": "foo",
		"y": "bar",
	}, args)
}

func TestSanitizerPreventsAutoEscape(t *testing.T) {
	// if a pipeline ends with a registered sanitizer name, the escaper must not append the default escape
	tpl := New("q").Sanitizers(SanitizerMap{
		"safe": func() Sanitizer { return &dummyX{} },
	})

	_, err := tpl.Parse(`SELECT {{ .Val | safe }}`)
	require.NoError(t, err)

	query, args, err := tpl.Execute(map[string]any{"Val": 42})
	require.NoError(t, err)

	// dummyX returns @x; if escaper wrongly appended an extra escape func we'd see another placeholder
	require.Equal(t, "SELECT @x", query)
	require.Equal(t, pgx.NamedArgs{"x": 42}, args)
}

func TestIfElseBranchEscapingAndArgs(t *testing.T) {
	tpl := New("q")
	_, err := tpl.Parse(`SELECT
{{ if .UseID }}
  WHERE id = {{ .ID }}
{{ else }}
  WHERE name = {{ .Name }}
{{ end }}`)
	require.NoError(t, err)

	q1, args1, err := tpl.Execute(map[string]any{"UseID": true, "ID": 7})
	require.NoError(t, err)
	require.Contains(t, q1, "WHERE id = @sqlt0")
	require.Equal(t, pgx.NamedArgs{"sqlt0": 7}, args1)

	q2, args2, err := tpl.Execute(map[string]any{"UseID": false, "Name": "alice"})
	require.NoError(t, err)
	require.Contains(t, q2, "WHERE name = @sqlt0")
	require.Equal(t, pgx.NamedArgs{"sqlt0": "alice"}, args2)
}

func TestRangeLoopEscaping(t *testing.T) {
	tpl := New("q")
	_, err := tpl.Parse(`VALUES {{ range $i, $v := .Items }}({{ $i }}, {{ $v }}) {{ end }}`)
	require.NoError(t, err)

	items := []any{"one", "two"}
	q, args, err := tpl.Execute(map[string]any{"Items": items})
	require.NoError(t, err)

	// Query should contain two placeholders for the two $v outputs.
	require.Contains(t, q, "@sqlt0")
	require.Contains(t, q, "@sqlt1")

	// args should have two entries (sqlt0 -> "one", sqlt1 -> "two")
	require.Equal(t, pgx.NamedArgs{
		"sqlt0": 0,
		"sqlt1": "one",
		"sqlt2": 1,
		"sqlt3": "two",
	}, args)
}

func TestWithBranchEscaping(t *testing.T) {
	tpl := New("q")
	_, err := tpl.Parse(`{{ with .User }}name={{ .Name }}{{ end }}`)
	require.NoError(t, err)

	q, args, err := tpl.Execute(map[string]any{"User": map[string]any{"Name": "bob"}})
	require.NoError(t, err)

	require.Contains(t, q, "name=@sqlt0")
	require.Equal(t, pgx.NamedArgs{"sqlt0": "bob"}, args)
}

func TestActionDeclarationBehavior(t *testing.T) {
	// Declaration inside an action should not itself cause output; later reference should be escaped.
	tpl := New("q")
	_, err := tpl.Parse(`{{$x := .Val}}{{$x}}`)
	require.NoError(t, err)

	q, args, err := tpl.Execute(map[string]any{"Val": "DECL"})
	require.NoError(t, err)

	// Only the second action produces output and it should be escaped to @sqlt0
	require.Equal(t, "@sqlt0", strings.TrimSpace(q))
	require.Equal(t, pgx.NamedArgs{"sqlt0": "DECL"}, args)
}

func TestSubTemplatesAndComments(t *testing.T) {
	// ensure that nested templates and comments are handled by the escaper and do not leak
	tpl := New("main")
	_, err := tpl.Parse(`
{{ define "inner" }}
-- comment inside template
WHERE col = {{ .Val }}
{{ end }}

SELECT * FROM t {{ template "inner" . }}
`)
	require.NoError(t, err)

	q, args, err := tpl.Execute(map[string]any{"Val": "v"})
	require.NoError(t, err)

	require.Contains(t, q, "WHERE col = @sqlt0")
	require.Equal(t, pgx.NamedArgs{"sqlt0": "v"}, args)
	require.Contains(t, q, "-- comment inside template")
}
