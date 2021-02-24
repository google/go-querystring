# go-querystring #

[![Go Reference](https://pkg.go.dev/badge/github.com/google/go-querystring/query.svg)](https://pkg.go.dev/github.com/google/go-querystring/query)
[![Test Status](https://github.com/google/go-querystring/workflows/tests/badge.svg)](https://github.com/google/go-querystring/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/google/go-querystring/branch/master/graph/badge.svg)](https://codecov.io/gh/google/go-querystring)

go-querystring is a Go library for encoding structs into URL query parameters.

## Usage ##

```go
import "github.com/google/go-querystring/query"
```

go-querystring is designed to assist in scenarios where you want to construct a
URL using a struct that represents the URL query parameters.  You might do this
to enforce the type safety of your parameters, for example, as is done in the
[go-github][] library.

The query package exports a single `Values()` function.  A simple example:

``` go
type opt struct {
  Query   string `url:"q"`
  ShowAll bool   `url:"all"`
  Page    int    `url:"page"`
}

v, _ := query.Values(opt{"foo", true, 2})
fmt.Print(v.Encode()) // will output: "q=foo&all=true&page=2"
```

### Supported types and options ###

The [package godocs][] are the authoritative source for documentation on
supported types and formatting options, but illustrative examples are provided
here as well.

#### booleans

By default, boolean values are encoded as the words "true" or "false":

``` go
type opt struct {
  V bool `url:"v"`
}

query.Values(opt{true}) // result: "v=true"
```

Adding the `int` option causes the field to be encoded as a "1" or "0":

``` go
type opt struct {
  V bool `url:v,int`
}

query.Values(opt{false}) // result: "v=0"
```

#### time

By default, time values are encoded as RFC3339 timestamps:

``` go
type opt struct {
  V time.Time `url:"v"`
}

query.Values(opt{false}) // result: "v=0"
```

Adding the `unix` option encodes as a UNIX timestamp (seconds since Jan 1, 1970)

``` go
type opt struct {
  V time.Time `url:"v,unix"`
}

query.Values(opt{false}) // result: "v=0"
```


[go-github]: https://github.com/google/go-github/commit/994f6f8405f052a117d2d0b500054341048fbb08
[package godocs]: https://pkg.go.dev/github.com/google/go-querystring/query

## Alternatives ##

If you are looking for a library that can both encode and decode query strings,
you might consider one of these alternatives:

 - https://github.com/gorilla/schema
 - https://github.com/pasztorpisti/qs
 - https://github.com/hetiansu5/urlquery
