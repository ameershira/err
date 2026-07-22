![Banner](./banner.png)
# err

[![Go Reference](https://pkg.go.dev/badge/github.com/ameershira/err.svg)](https://pkg.go.dev/github.com/ameershira/err)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ameershira/err)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/ameershira/err/go.yml)
![GitHub](https://img.shields.io/github/license/ameershira/err)

`err` is a small Go error-handling helper package that produces a **single `error` value enriched with call-site context**
(file, line number, and function name), while remaining fully compatible with Go’s standard `errors` APIs.

It is designed for **transparent, low-friction error wrapping** without custom error types or hidden behavior.

---

## Features

- Automatically includes **file, line number, and function name**
- Supports **multiple wrapped errors** via `errors.Join`
- Argument-order aware and convention-friendly
- Fully compatible with:
  - `errors.Is`
  - `errors.As`
  - `errors.AsType`
  - `errors.Join`
- No custom error types
- No global state
- Zero runtime configuration

---

## Design Goals

This package exists to solve a very specific problem:

> *“I want structured, contextual errors without inventing a new error system.”*

Key principles:

- **One error interface** – always returns `error`
- **Explicit wrapping** – no magic stack traces
- **Predictable ordering** – error chains read top-to-bottom
- **Standard tooling first** – plays nicely with Go’s stdlib

---

## Installation
```sh
go get github.com/ameershira/err
```

---

## Usage

Import the package (optionally dot-imported for brevity):

```go
import e "github.com/ameershira/err"
```

---

## Err

`Err` is the primary API. It accepts a mix of `error`, `string`, and `nil` values and returns a single contextual error.

### Supported Forms

```go
Err(error)
Err(error1, error2)
Err(error, "msg")
Err("msg")
Err(error1, error2, error3, "msg")
Err()
Err(nil)                             // returns nil
Err("%s %d", "hello", 100)           // formatted strings
Err(error1, error2, "%s %d", "hello", 100)
```
Note that a string should be the last error argument. If a string has remaining arguments after it, those arguments are used as printf-style formatting arguments.

### Example

```go
if err := doSomething(); err != nil {
    return e.Err(err, "failed to do something")
}
```

This produces an error chain similar to:

```
path/file.go:42: myFunc: failed to do something
path/file2.go:30: doSomething: file A does not exist
```

---

## Argument Ordering (Important)

By convention:

1. **Error values come first**
2. **Text messages come last**

```go
Err(err, "custom message")
```

`Err("custom message", err)` does not wrap `err`; once `Err` sees a string, all remaining arguments are treated as formatting arguments for that string.

Following the convention ensures:

- Readable error chains
- Consistency with `errors.Join`
- Fewer mistakes when mixing APIs

## Nil Handling

- `Err(nil)` → `nil`
- `Err(nil, nil)` → `nil`
- `Err(err, nil)` → contextual error containing `err`
- `Err(nil, "msg")` → contextual error with message
- `Err(nil, "%s %d", "hello", 100)` → contextual error with formatted message
- `Err()` → contextual error (trace only)

This makes it safe to use in early returns without extra checks.

---

## Logging & Tracing

- Call-site information is captured via `runtime.Caller`
- Internal failures are logged using `log/slog`
- The returned error remains valid even if tracing fails

---

## When *Not* to Use This

Use `errors.Join` directly if:

- You **do not want call-site context**
- You are aggregating errors at a boundary (e.g. validation)

This package intentionally does **not** replace Go’s standard error APIs—it complements them.

---

## Philosophy

> Errors should be:
> - Easy to create
> - Hard to misuse
> - Obvious when printed
> - Boring to maintain

This package aims to keep error handling **mechanical, explicit, and readable**, especially in larger codebases where error context matters.
