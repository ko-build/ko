# `golicenses`

[![Go Reference](https://pkg.go.dev/badge/github.com/imjasonh/golicenses.svg)](https://pkg.go.dev/github.com/imjasonh/golicenses)
[![Update](https://github.com/imjasonh/golicenses/actions/workflows/update.yaml/badge.svg)](https://github.com/imjasonh/golicenses/actions/workflows/update.yaml)

This is an **experimental** package to lookup the license for a Go package.

This is not guaranteed to work, to update regularly, or to continue to have the same API.
At a minimum, I'll probably change the repo name if I can think of something better.

For example:

```golang
lic, _ := golicenses.Get("github.com/google/go-containerregistry")
fmt.Println(lic)
```

prints

```
Apache-2.0
```

This is based on the public BigQuery dataset provided by https://deps.dev/.
See [How are licenses determined?](https://deps.dev/faq#how-are-licenses-determined) for more information.

This repo periodically queries the public dataset and regenerates `licenses.csv`, which is gzipped and `//go:embed`ed into the package.

The result is a ~3MB dependency that can be loaded and queried in ~200ms the first time -- subsequent calls take microseconds.

There are almost certainly more optimizations that could improve both size and query time. PRs welcome!
