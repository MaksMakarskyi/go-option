![Go Version](https://img.shields.io/badge/go-1.27+-blue)
![OS](https://img.shields.io/badge/OS-Windows,Linux,MacOS-orange)
![License](https://img.shields.io/badge/license-MIT-green)

# Table of Content

- [About](#about)
- [Install](#install)
- [Quick start](#quick-start)
- [Examples](#examples)
- [Reference](#reference)
- [Gotchas](#gotchas)

# About

This is the implementation of the Option type for the Go programming language. It was greatly inspired by Rust's
`Option<T>` and the experience of using it. Go has three ways to say "this might not be there", and each one might be 
inconvenient in one way or another:
- `*T` — nil panics, and you can't write `&3`. Moreover, it forces you to manually check for the nil pointer all over
the place, resulting in a lot of `if` statements.
- `(T, bool)` — can't be stored in a struct, can't be chained. Again, requires you to write additional `if` statements.
- the zero value — can't tell `0` from "not set".  Usually, one would use pointers, which have their own problems and inconveniences.

`Option[T]` says it once, in the type. No allocation, no nil, and the zero value is already a valid `None`. It also implements the following interfaces:
- `fmt.Formatter`
- `driver.Valuer` / `sql.Scanner`
- `json.Marshaler` / `json.Unmarshaler`
- `encoding.TextMarshaler` / `encoding.TextUnmarshaler`

> [!NOTE]
> This package requires **Go 1.27+** because it uses generic methods.

# Install

```bash
go get github.com/MaksMakarskyi/go-option
```

Import:

```go
import option "github.com/MaksMakarskyi/go-option"
```

# Quick start

```go
a := option.Some(42)
b := option.None[int]()

a.IsSome()          // true
a.Unwrap()          // 42
b.UnwrapOr(7)       // 7
b.Get()             // 0, false

a.Filter(func(n int) bool { return n > 100 })  // None
a.Map(strconv.Itoa)                            // Some("42")
```

# Examples

## 1. JSON fields that can be zero

`omitempty` deletes real data. A user who sets `retries` to `0` gets the same JSON as a user who didn't set it at all:

```go
type Settings struct {
    Retries int  `json:"retries,omitempty"`
    Debug   bool `json:"debug,omitempty"`
}

json.Marshal(Settings{Retries: 0, Debug: false})  // {}  ← both silently lost
```

The usual fix is `*int` and `*bool`, which means temp variables everywhere (`&0` doesn't compile) and a nil check on every read. With `Option` plus `omitzero`:

```go
type Settings struct {
    Retries option.Option[int]  `json:"retries,omitzero"`
    Debug   option.Option[bool] `json:"debug,omitzero"`
}

json.Marshal(Settings{Retries: option.Some(0), Debug: option.Some(false)})
// {"retries":0,"debug":false}

json.Marshal(Settings{})
// {}

var s Settings
json.Unmarshal([]byte(`{"retries":0}`), &s)
s.Retries.UnwrapOr(3)  // 0   — set, and kept
s.Debug.IsSome()       // false — never sent
```

| | `0` set | not set | reading it |
|---|---|---|---|
| `int` + `omitempty` | ❌ dropped | ✅ dropped | – |
| `*int` + `omitempty` | ✅ `0` | ✅ dropped | nil check |
| `Option[int]` + `omitzero` | ✅ `0` | ✅ dropped | `UnwrapOr` |

## 2. Lookups that can fail at every step

Three lookups, three ways to come up empty, one sentinel repeated four times:

```go
func managerName(id string) string {
    u, ok := users[id]
    if !ok {
        return "unassigned"
    }
    if u.ManagerID == nil {
        return "unassigned"
    }
    m, ok := users[*u.ManagerID]
    if !ok {
        return "unassigned"
    }
    return m.Name
}
```

`AndThen` short-circuits, so the empty case is written once:

```go
func findUser(id string) option.Option[User] {
    u, ok := users[id]
    return option.FromTuple(u, ok)
}

func managerName(id string) string {
    return findUser(id).
        AndThen(func(u User) option.Option[string] { return u.ManagerID }).
        AndThen(findUser).
        Map(func(u User) string { return u.Name }).
        UnwrapOr("unassigned")
}
```

## 3. Defaults and validation

```go
// before
timeout := 30
if cfg.Timeout != nil {
    timeout = *cfg.Timeout
}

port := 8080
if cfg.Port != nil && *cfg.Port > 0 {
    port = *cfg.Port
}

// after
timeout := cfg.Timeout.UnwrapOr(30)
port := cfg.Port.Filter(func(p int) bool { return p > 0 }).UnwrapOr(8080)
```

# Reference

**Create**

| Function/Method | Description |
|---|---|
| `Some(v)` | an Option holding `v` |
| `None[T]()` | an empty Option |
| `FromTuple(v, ok)` | wraps a comma-ok result |
| `FromPtr(p)` | `nil` → `None`, otherwise a copy of `*p` |
| `o.ToPtr()` | pointer to a copy, or `nil` |

**Check**

| Function/Method | Description |
|---|---|
| `o.IsSome()` / `o.IsNone()` | is a value present |
| `o.IsSomeAnd(pred)` | present *and* matches `pred` |

**Read**

| Function/Method | Description |
|---|---|
| `o.Get()` | `(value, ok)` |
| `o.Unwrap()` | value, or panic |
| `o.Expect(msg)` | value, or panic with `msg` |
| `o.UnwrapOr(v)` | value, or `v` |
| `o.UnwrapOrElse(fn)` | value, or `fn()` |
| `o.UnwrapOrZero()` | value, or the zero value of `T` |
| `o.OkOr(err)` | `(value, err)`, wrapped so `errors.Is` matches `ErrValueIsNone` too |
| `o.OkOrElse(fn)` | `(value, fn())` — your error, untouched |

**Transform**

| Function/Method | Description |
|---|---|
| `o.Map(fn)` | apply `fn` to the value |
| `o.MapOr(v, fn)` | `fn(value)`, or `v` |
| `o.AndThen(fn)` | chain a step that may be empty |
| `o.Filter(pred)` | keep the value only if it matches |

**Combine**

| Function/Method | Description |
|---|---|
| `o.And(b)` | `b` if `o` is `Some` |
| `o.Or(b)` / `o.OrElse(fn)` | first `Some` |
| `o.Xor(b)` | `Some` only if exactly one is |

Callbacks only run when they have something to do: `IsSomeAnd`, `Map`, `MapOr`, `AndThen` and
`Filter` skip `fn` on a `None`, while `UnwrapOrElse`, `OrElse` and `OkOrElse` call it *only* on a `None`.

**Interfaces**

| Interface | On a `Some` | On a `None` |
|---|---|---|
| `fmt.Formatter` | prints as the value would | prints `None` |
| `json.Marshaler` / `json.Unmarshaler` | the value | `null` |
| `encoding.TextMarshaler` / `encoding.TextUnmarshaler` | the value as text | empty text |
| `driver.Valuer` / `sql.Scanner` | the value | SQL `NULL` |

# Gotchas

**1. Use `omitzero`, not `omitempty`.** `omitempty` has never worked on structs and will emit `null`. `omitzero` (Go 1.24+) works because `None` *is* the zero value. It is also the future-proof choice: under `encoding/json/v2` `omitempty` does drop a `None`, but it drops a deliberate `Some("")` along with it.

**2. `FromTuple(m[k])` doesn't compile.** Comma-ok is only special in an assignment. Either assign first, or wrap a function that returns `(T, bool)`.

**3. `==` works when `T` is comparable.** `Some(5) == Some(5)` is true. For a `T` that isn't comparable (a slice, a map), use `Get()`.

**4. Options are values.** Copying one copies the value inside, and `ToPtr` hands back a copy, so nothing can reach in and mutate an Option after the fact.

**5. In text, `None` and `Some("")` are identical.** Text has no `null`, so `MarshalText` writes empty for both and reads empty back as `None`. This only bites a `T` whose text can be empty — `string` and `[]byte`. If you need the distinction, keep the Option a JSON *value*, where `null` is available, rather than a map key.

**6. SQL errors surface at run time, not compile time.** `Value()` reports an error for a `T` the driver cannot store (`[]int`, a struct). It cannot be a compile error, because a method may not narrow its receiver's type parameter. A `None` is always `NULL` whatever `T` is, and values are converted on the way out — an `Option[int]` sends an `int64`.

**7. An Option prints as its value.** `Format` forwards the verb, so `%v` of `Some(42)` is `42`, not `{42 true}`, and a `None` is `None` under every verb. `%T` still reports `option.Option[int]`.

**8. `Unwrap` and `Expect` panic.** They are for values you know are there; everywhere else reach for `Get`, `UnwrapOr` or `OkOr`. `Expect(msg)` returns the value and panics with `msg`, matching Rust's.
