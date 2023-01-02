# Integrating Policy Verification

The goal of this package is to make it easy for downstream tools to incorporate
the verification capabilities of `ClusterImagePolicy` in other contexts where
OCI artifacts are consumed.

The most straightforward example of this is to enable OCI build tooling to
incorporate policies over the base images on top of which an application image
is built (e.g. `ko`, `kaniko`).  However, this can be used by other tooling
that stores artifacts in OCI registries to verify those as well, examples of
this could include the way Buildpacks v3 and Crossplane store elements in OCI
registries.

## Configuration

Verification is configured via `policy.Verification`:

```golang
type Verification struct {
	// NoMatchPolicy specifies the behavior when a base image doesn't match any
	// of the listed policies.  It allows the values: allow, deny, and warn.
	NoMatchPolicy string `yaml:"no-match-policy,omitempty"`

	// Policies specifies a collection of policies to use to cover the base
	// images used as part of evaluation.  See "policy" below for usage.
	// Policies can be nil so that we can distinguish between an explicitly
	// specified empty list and when policies is unspecified.
	Policies *[]Source `yaml:"policies,omitempty"`
}
```

`NoMatchPolicy` controls the behavior when an image reference is passed that
does not match any of the configured policies.

`Policies` can be specified via three possible sources:

```golang
// Source contains a set of options for specifying policies.  Exactly
// one of the fields may be specified for each Source entry.
type Source struct {
	// Data is a collection of one or more ClusterImagePolicy resources.
	Data string `yaml:"data,omitempty"`

	// Path is a path to a file containing one or more ClusterImagePolicy
	// resources.
	// TODO(mattmoor): Make this support taking a directory similar to kubectl.
	// TODO(mattmoor): How do we want to handle something like -R?  Perhaps we
	// don't and encourage folks to list each directory individually?
	Path string `yaml:"path,omitempty"`

	// URL links to a file containing one or more ClusterImagePolicy resources.
	URL string `yaml:"url,omitempty"`
}
```

### With `spf13/viper`

Many tools leverage `spf13/viper` for configuration, and `policy.Verification`
may be used in conjunction with viper via:

```golang
	vfy := policy.Verification{}
	if err := v.UnmarshalKey("verification", &vfy); err != nil { ... }
```

This allows a section of the viper config:

```yaml
verification:
  noMatchPolicy: deny
  policies:
  - data: ... # Inline policies
  - url: ... # URL to policies
  ...
```

## Compilation

The `policy.Verification` can be compiled into a `policy.Verifier` using
`policy.Compile`, which also takes a `context.Context` and a function that
controls how warnings are surfaced:

```golang
	verifier, err := policy.Compile(ctx, verification,
		func(s string, i ...interface{}) {
			// Handle warnings your own way!
		})
	if err != nil { ... }
```

The compilation process will surface compilation warnings via the supplied
function and return any errors resolving or compiling the policies immediately.

## Verification

With a compiled `policy.Verifier` many image references can be verified against
the compiled policies by invoking `Verify`:
```golang
// Verifier is the interface for checking that a given image digest satisfies
// the policies backing this interface.
type Verifier interface {
	// Verify checks that the provided reference satisfies the backing policies.
	Verify(context.Context, name.Reference, authn.Keychain) error
}
```
