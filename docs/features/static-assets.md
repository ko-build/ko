# Static Assets

`ko` can also bundle static assets into the images it produces.

By convention, any contents of a directory named `<importpath>/kodata/` will be
bundled into the image, and the path where it's available in the image will be
identified by the environment variable `KO_DATA_PATH`.

As an example, you can bundle and serve static contents in your image:

```
cmd/
  app/
    main.go
    kodata/
      favicon.ico
      index.html
```

Then, in your `main.go`:

```go
func main() {
    http.Handle("/", http.FileServer(http.Dir(os.Getenv("KO_DATA_PATH"))))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

You can simulate `ko`'s behavior outside of the container image by setting the
`KO_DATA_PATH` environment variable yourself with `KO_DATA_PATH=cmd/app/kodata/ go run ./cmd/app`.

> 💡 **Tip:** Symlinks in `kodata` are followed and included as well. For example,
you can include Git commit information in your image with `ln -s -r .git/HEAD ./cmd/app/kodata/`

Also note that `http.FileServer` will not serve the `Last-Modified` header
(or validate `If-Modified-Since` request headers) because `ko` does not embed
timestamps by default.

This can be supported by manually setting the `KO_DATA_DATE_EPOCH` environment
variable during build ([See FAQ](../../advanced/faq#why-are-my-images-all-created-in-1970)).

