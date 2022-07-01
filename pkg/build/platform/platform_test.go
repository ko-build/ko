package platform

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestMatchesPlatformSpec(t *testing.T) {
	for _, tc := range []struct {
		platform *v1.Platform
		spec     []string
		result   bool
		err      bool
	}{{
		platform: nil,
		spec:     []string{"all"},
		result:   true,
	}, {
		platform: nil,
		spec:     []string{"linux/amd64"},
		result:   false,
	}, {
		platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		spec:   []string{"all"},
		result: true,
	}, {
		platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "windows",
		},
		spec:   []string{"linux"},
		result: false,
	}, {
		platform: &v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		spec:   []string{"linux/amd64", "linux/arm64"},
		result: true,
	}, {
		platform: &v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		spec:   []string{"linux/amd64", "linux/arm64/v4"},
		result: false,
	}, {
		platform: &v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		spec: []string{"linux/amd64", "linux/arm64/v3/z5"},
		err:  true,
	}, {
		spec: []string{},
		platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		result: false,
	}, {
		// Exact match w/ osversion
		spec: []string{"windows/amd64:10.0.17763.1234"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: true,
	}, {
		// OSVersion partial match using relaxed semantics.
		spec: []string{"windows/amd64:10.0.17763"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: true,
	}, {
		// Not windows and osversion isn't exact match.
		spec: []string{"linux/amd64:10.0.17763"},
		platform: &v1.Platform{
			OS:           "linux",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: false,
	}, {
		// Not matching X.Y.Z
		spec: []string{"windows/amd64:10"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: false,
	}, {
		// Requirement is more specific.
		spec: []string{"windows/amd64:10.0.17763.1234"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763", // this won't happen in the wild, but it shouldn't match.
		},
		result: false,
	}, {
		// Requirement is not specific enough.
		spec: []string{"windows/amd64:10.0.17763.1234"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234.5678", // this won't happen in the wild, but it shouldn't match.
		},
		result: false,
	}} {
		pm, err := ParseSpec(tc.spec)
		if tc.err {
			if err == nil {
				t.Errorf("parseSpec(%v, %q) expected err", tc.platform, tc.spec)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseSpec failed for %v %q: %v", tc.platform, tc.spec, err)
		}
		matches := pm.Matches(tc.platform)
		if got, want := matches, tc.result; got != want {
			t.Errorf("wrong result for %v %q: want %t got %t", tc.platform, tc.spec, want, got)
		}
	}
}

func TestIsWasi(t *testing.T) {
	type args struct {
		p *v1.Platform
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "wasi platform",
			args: args{
				p: &v1.Platform{
					OS:           "wasm",
					Architecture: "wasi",
				},
			},
			want: true,
		},
		{
			name: "not wasi Architecture",
			args: args{
				p: &v1.Platform{
					OS:           "wasm",
					Architecture: "foo",
				},
			},
			want: false,
		},
		{
			name: "not wasm os",
			args: args{
				p: &v1.Platform{
					OS:           "bar",
					Architecture: "wasi",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, want := IsWasi(tt.args.p), tt.want
			if got != want {
				t.Errorf("wrong result for %v: want %v got %v", tt.args.p, want, got)
			}
		})
	}
}
