package starlark

import (
	"context"
	"errors"
	"math/rand"

	"github.com/DLag/starlight/convert"
	"go.starlark.net/starlark"
)

func NewRand(ctx context.Context) starlark.Value {
	return NewBuiltIn(ctx, map[string]builtinGetter{"int": randInt})
}

func randInt(ctx context.Context) *starlark.Builtin {
	l := logger(ctx)
	return starlark.NewBuiltin("rand_int", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		l := l.With().Str("starlark_function", "var_get").Logger()
		if args.Len() != 1 || args.Index(0).Type() != "int" {
			l.Error().Int("args_len", args.Len()).Msg("Wrong args. Should be rand_int(int).")
			return starlark.None, errors.New("wrong args, should be rand_int(int)")
		}
		max, _ := args.Index(0).(starlark.Int).Int64()
		return convert.ToValue(rand.Int63n(max))
	})
}
