package sqlt

import (
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
)

type argFormatter struct {
	args []any
}

func (f *argFormatter) Format(arg any) string {
	idx := slices.Index(f.args, arg)
	if idx == -1 {
		idx = len(f.args)
		f.args = append(f.args, arg)
	}
	return fmt.Sprintf("@sqlt%d", idx)
}

func (f *argFormatter) Args() pgx.NamedArgs {
	out := make(pgx.NamedArgs, len(f.args))
	for i, v := range f.args {
		out[fmt.Sprintf("sqlt%d", i)] = v
	}
	return out
}
