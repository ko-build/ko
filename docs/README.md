# Docs for https://ko.build

## Development

Update `.md` files to update content.

Update `mkdocs.yml` to update sidebar headers and ordering.

To run locally:

- [install `mkdocs` and `mkdocs-material`](https://squidfunk.github.io/mkdocs-material/getting-started/) and run `mkdocs serve`, or
- `docker run --rm -it -p 8000:8000 -v ${PWD}:/docs squidfunk/mkdocs-material`
  - on an M1 Mac, use `ghcr.io/afritzler/mkdocs-material` instead.

This will start a local server on localhost:8000 that autoupdates as you make changes.

## Deployment

When PRs are merged, the site will be rebuilt and published automatically.

### Credits

The site is powered by [mkdocs-material](https://squidfunk.github.io/mkdocs-material). The code and theme are released under the MIT license.

Content is licensed [CC-BY](https://creativecommons.org/licenses/by/4.0/).

