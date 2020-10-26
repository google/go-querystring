# go-querystring #

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/google/go-querystring/query)
[![Test Status](https://github.com/google/go-querystring/workflows/tests/badge.svg)](https://github.com/google/go-querystring/actions?query=workflow%3Atests)

go-querystring is Go library for encoding structs into URL query parameters.

## Usage ##

```go
import "github.com/google/go-querystring/query"
```

go-querystring is designed to assist in scenarios where you want to construct a
URL using a struct that represents the URL query parameters.  You might do this
to enforce the type safety of your parameters, for example, as is done in the
[go-github][] library.

The query package exports a single `Values()` function.  A simple example:

```go
type Options struct {
  Query   string `url:"q"`
  ShowAll bool   `url:"all"`
  Page    int    `url:"page"`
}

opt := Options{ "foo", true, 2 }
v, _ := query.Values(opt)
fmt.Print(v.Encode()) // will output: "q=foo&all=true&page=2"
```

[go-github]: https://github.com/google/go-github/commit/994f6f8405f052a117d2d0b500054341048fbb08

## License ##

This library is distributed under the BSD-style license found in the [LICENSE](./LICENSE)
file.
