# How to Contribute to ko

We'd love to accept your patches and contributions to this project. There are
just a few small guidelines you need to follow.

## Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

## Testing

Ensure the following passes:
```
./hack/presubmit.sh
```
and commit any resultant changes to `go.mod` and `go.sum`. To update any docs
after client changes, run:

```
./hack/update-codegen.sh
```
