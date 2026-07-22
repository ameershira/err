// err is a general error handling package that returns a
// single error interface which includes the file, line number,
// function name, an optional error msg, and optional wrapped errors.
package err

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"strings"
)

// Err handles errors with optional messages, supporting multiple errors and simple message wrapping.
// The cases below are supported:
//  1. Err(error)
//  2. Err(error1, error2)
//  3. Err(error, "msg")
//  4. Err("msg")
//  5. Err(error1, error2, error3, "msg")
//  6. Err()
//  7. Err(nil) // This returns nil.
//  8. Err("%s %d", "hello", 100) // printf format
//  9. Err(error1, error2, "%s %d", "hello", 100)
//
// The order of the arguments matters. All error types are wrapped in order of
// the positional arguments. The first string or string-like argument is treated
// as the message or printf-style format string, and all remaining arguments are
// used as formatting arguments.
//
// This is important as traces of errors need to be wrapped in a specific order
// to make sense when printed or logged as a stack trace.
//
// By convention we follow this format:
// The error types are positioned first, then the string error message.
//
//	import e "switchman-srv/pkg/err"
//
//	if err := funcA(); err != nil {
//	    return e.Err(err, "my custom msg")
//	}
//
// Stick to this convention because it is the same as errors.Join(). This reduces
// error handling mistakes when errors.Join() and other standard error APIs are
// used alongside this API.
//
// The Err() API always adds a callsite trace. For cases where you do not want a trace, use errors.Join().
func Err(args ...any) error {
	if len(args) == 1 && args[0] == nil {
		return nil
	}

	var funcName string

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		slog.Error("Err()", "err", "runtime.Caller(1) failed")
	} else {
		// Get the calling function's name
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			slog.Error("Err()", "err", "internal err pkg error: runtime.FuncForPC(pc) failed")
		} else {
			funcName = trimFuncName(fn.Name())
		}
	}

	return parseErrors(line, file, funcName, args...)
}

// parseErrors constructs a combined error from a set of mixed arguments.
// It accepts a source line number, file name, and function name for context,
// followed by variadic arguments of type string, string-like, error, nil, or
// formatting values after a string.
//
// The first string or string-like argument is formatted into a new error
// message of the form:
//
//	"<file>:<line>: <funcName>: <message>"
//
// Error arguments before the first string are included as-is, while nil values
// are ignored. Arguments after the first string are consumed as formatting
// values.
//
// If no string message is provided, a default contextual error is appended
// (e.g. "<file>:<line>: <funcName>").
//
// If all arguments are nil, parseErrors returns nil. If any invalid argument
// type is encountered, it logs an error using slog and continues processing.
//
// The returned error combines all collected errors using errors.Join.
func parseErrors(line int, file, funcName string, args ...any) error {
	var errs, stringErrs []error
	var msgFound, nilFound bool

forloop:
	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			msgFound = true
			// a plain string should be the last element, if not then it is a printf format string with args that follow.
			fmtMsg := fmt.Sprintf(v, args[i+1:]...)
			stringErrs = append(stringErrs, fmt.Errorf("%s:%d: %s: %s", file, line, funcName, fmtMsg))
			break forloop
		case error:
			errs = append(errs, v)
		case nil:
			// ignore
			nilFound = true
		default:
			rv := reflect.ValueOf(v)
			// Check if the underlying concrete type is string.
			if rv.Kind() == reflect.String {
				msgFound = true
				// a plain string should be the last element, if not then it is a printf format string with args that follow.
				fmtMsg := fmt.Sprintf(rv.String(), args[i+1:]...)
				stringErrs = append(stringErrs, fmt.Errorf("%s:%d: %s: %s", file, line, funcName, fmtMsg))
				break forloop
			}

			slog.Error("Err()", "err", fmt.Sprintf("invalid argument type: expected string or error, got %s at %s:%d [%s]", reflect.TypeOf(arg), file, line, funcName))
		}
	}

	if nilFound && len(errs) == 0 && len(stringErrs) == 0 {
		return nil
	}

	// if no error string msg was received then add a default to the tail of the slice
	if !msgFound {
		err := fmt.Errorf("%s:%d: %s", file, line, funcName)
		errs = append(errs, err)
	} else {
		// Errors created via string messges are added to the tail of the slice
		errs = append(errs, stringErrs...)
	}

	return errors.Join(errs...)
}

// trimFuncName removes unnecessary package path details.
func trimFuncName(fullName string) string {
	parts := strings.Split(fullName, "/")
	pLen := len(parts)
	if pLen == 1 {
		return parts[0]
	}
	return parts[len(parts)-1] // Keep only the last part (func name)
}
