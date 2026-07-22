package err_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	e "github.com/ameershira/err"
)

type customErr1 struct{}

func (c customErr1) Error() string { return "custom error type1" }

type customErr2 struct{}

func (c customErr2) Error() string { return "custom error type2" }

type customString string

var (
	custErr1  customErr1
	custErr2  customErr2
	customStr customString = "custom string type"

	stdErr1 = errors.New("std err1")
	stdErr2 = errors.New("std err2")
	stdErr3 = errors.New("std err3")
)

type testCase struct {
	in  []any
	out []any
}

var testCases = []testCase{
	{in: []any{}, out: []any{}}, // Empty out just checks callsite context

	{in: []any{nil}, out: []any{nil}},
	{in: []any{nil, nil}, out: []any{nil}},
	{in: []any{nil, nil, nil}, out: []any{nil}},

	{in: []any{stdErr1}, out: []any{stdErr1}},
	{in: []any{stdErr1, stdErr2}, out: []any{stdErr1, stdErr2}},
	{in: []any{stdErr1, stdErr2, stdErr3}, out: []any{stdErr1, stdErr2, stdErr3}},
	{in: []any{nil, stdErr1, nil}, out: []any{stdErr1}},

	{in: []any{custErr1}, out: []any{custErr1}},
	{in: []any{custErr1, custErr2}, out: []any{custErr1, custErr2}},

	{in: []any{"errmsg"}, out: []any{"errmsg"}},
	{in: []any{nil, "errmsg"}, out: []any{"errmsg"}},
	{in: []any{nil, nil, nil, "errmsg"}, out: []any{"errmsg"}},
	{in: []any{stdErr1, "errmsg"}, out: []any{stdErr1, "errmsg"}},
	{in: []any{stdErr1, stdErr2, stdErr3, "errmsg"}, out: []any{stdErr1, stdErr2, stdErr3, "errmsg"}},

	{in: []any{customStr}, out: []any{customStr}},
	{in: []any{nil, customStr}, out: []any{customStr}},

	{in: []any{"hello_%s_%d", "world", 100}, out: []any{"hello_world_100"}},
	{in: []any{nil, "hello_%s_%d", "world", 100}, out: []any{"hello_world_100"}},
	{in: []any{nil, nil, nil, "hello_%s_%d", "world", 100}, out: []any{"hello_world_100"}},
	{in: []any{stdErr1, "hello_%s_%d", "world", 100}, out: []any{stdErr1, "hello_world_100"}},
	{in: []any{stdErr1, stdErr2, stdErr3, "hello_%s_%d", "world", 100}, out: []any{stdErr1, stdErr2, stdErr3, "hello_world_100"}},
}

func generateTestName(tc testCase) string {
	var args []string

argsLoop:
	for i, arg := range tc.in {
		switch arg.(type) {
		case nil:
			args = append(args, "nil")
		case customString:
			args = append(args, "customString")
		case string:
			if i < len(tc.in)-1 {
				args = append(args, "format", "args")
				break argsLoop
			}
			args = append(args, "string")
		case customErr1:
			args = append(args, "customErr1")
		case customErr2:
			args = append(args, "customErr2")
		case error:
			args = append(args, "error")
		default:
			args = append(args, "unknown")
		}
	}

	return fmt.Sprintf("Err(%s)", strings.Join(args, ","))
}

func TestErr(t *testing.T) {
	for _, tc := range testCases {
		t.Run(generateTestName(tc), func(t *testing.T) {

			err := e.Err(tc.in...)

			wantNil := false
			for _, output := range tc.out {
				switch v := output.(type) {
				case nil:
					wantNil = true
					if err != nil {
						t.Errorf("got %q, want nil", err)
					}
				case customString:
					if err == nil {
						t.Errorf("got nil, want error containing %q", v)
						return
					}
					if !strings.Contains(err.Error(), string(v)) {
						t.Errorf("got %q, want error containing %q", err, v)
					}
				case string:
					if err == nil {
						t.Errorf("got nil, want error containing %q", v)
						return
					}
					if !strings.Contains(err.Error(), v) {
						t.Errorf("got %q, want error containing %q", err, v)
					}
				case customErr1:
					_, ok := errors.AsType[customErr1](err)
					if !ok {
						t.Errorf("errors.AsType failed: got %T(%q), want error chain containing %T", err, err, v)
					}

					var target customErr1
					if !errors.As(err, &target) {
						t.Errorf("errors.As failed: got %T(%q), want error chain containing %T", err, err, v)
					}
				case customErr2:
					_, ok := errors.AsType[customErr2](err)
					if !ok {
						t.Errorf("errors.AsType failed: got %T(%q), want error chain containing %T", err, err, v)
					}

					var target customErr2
					if !errors.As(err, &target) {
						t.Errorf("errors.As failed: got %T(%q), want error chain containing %T", err, err, v)
					}
				case error:
					if !errors.Is(err, v) {
						t.Errorf("got %q, want error wrapping %q", err, v)
					}

				default:
					t.Errorf("unsupported test output type %T", v)
				}
			}

			if wantNil {
				return
			}

			if err == nil {
				t.Errorf("got nil, want callsite context")
				return
			}

			// The expected call site context
			for _, want := range []string{"err_test.go", "TestErr"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("got %q, want error containing %q", err, want)
				}
			}
		})
	}
}

func TestErrWrapping(t *testing.T) {
	err1 := e.Err("errmsg1")
	err2 := e.Err("errmsg2")

	err := e.Err(err1, err2)

	for _, want := range []error{err1, err2} {
		if !errors.Is(err, want) {
			t.Errorf("got %q, want error wrapping %q", err, want)
		}
	}

	for _, want := range []string{"errmsg1", "errmsg2"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("got %q, want error containing %q", err, want)
		}
	}
}

func funcA() error { return e.Err(funcB()) }
func funcB() error { return e.Err(funcC(), "xyz") }
func funcC() error { return e.Err("root error!") }

func TestStackContext(t *testing.T) {
	err := funcA()

	for _, want := range []string{"funcC", "err_test.go", "root error!", "xyz"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("got %q, want error containing %q", err, want)
		}
	}

	wrapped := fmt.Errorf("wrappedA: %w", funcA())

	for _, want := range []string{"funcC", "wrappedA"} {
		if !strings.Contains(wrapped.Error(), want) {
			t.Errorf("got %q, want error containing %q", wrapped, want)
		}
	}
}
