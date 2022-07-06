package binary

import (
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)


func TestGoarm(t *testing.T) {
	// From golang@sha256:1ba0da74b20aad52b091877b0e0ece503c563f39e37aa6b0e46777c4d820a2ae
	// and made up invalid cases.
	for _, tc := range []struct {
		platform v1.Platform
		variant  string
		err      bool
	}{{
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "vnot-a-number",
		},
		err: true,
	}, {
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "wrong-prefix",
		},
		err: true,
	}, {
		platform: v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		variant: "",
	}, {
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "v5",
		},
		variant: "5",
	}, {
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "v7",
		},
		variant: "7",
	}, {
		platform: v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v8",
		},
		variant: "7",
	},
	} {
		variant, err := getGoarm(tc.platform)
		if tc.err {
			if err == nil {
				t.Errorf("getGoarm(%v) expected err", tc.platform)
			}
			continue
		}
		if err != nil {
			t.Fatalf("getGoarm failed for %v: %v", tc.platform, err)
		}
		if got, want := variant, tc.variant; got != want {
			t.Errorf("wrong variant for %v: want %q got %q", tc.platform, want, got)
		}
	}
}

func TestBuildEnv(t *testing.T) {
	tests := []struct {
		description  string
		platform     v1.Platform
		userEnv      []string
		configEnv    []string
		expectedEnvs map[string]string
	}{{
		description: "defaults",
		platform: v1.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		expectedEnvs: map[string]string{
			"GOOS":        "linux",
			"GOARCH":      "amd64",
			"CGO_ENABLED": "0",
		},
	}, {
		description: "override a default value",
		configEnv:   []string{"CGO_ENABLED=1"},
		expectedEnvs: map[string]string{
			"CGO_ENABLED": "1",
		},
	}, {
		description: "override an envvar and add an envvar",
		userEnv:     []string{"CGO_ENABLED=0"},
		configEnv:   []string{"CGO_ENABLED=1", "GOPRIVATE=git.internal.example.com,source.developers.google.com"},
		expectedEnvs: map[string]string{
			"CGO_ENABLED": "1",
			"GOPRIVATE":   "git.internal.example.com,source.developers.google.com",
		},
	}, {
		description: "arm variant",
		platform: v1.Platform{
			Architecture: "arm",
			Variant:      "v7",
		},
		expectedEnvs: map[string]string{
			"GOARCH": "arm",
			"GOARM":  "7",
		},
	}, {
		// GOARM is ignored for arm64.
		description: "arm64 variant",
		platform: v1.Platform{
			Architecture: "arm64",
			Variant:      "v8",
		},
		expectedEnvs: map[string]string{
			"GOARCH": "arm64",
			"GOARM":  "",
		},
	}, {
		description: "amd64 variant",
		platform: v1.Platform{
			Architecture: "amd64",
			Variant:      "v3",
		},
		expectedEnvs: map[string]string{
			"GOARCH":  "amd64",
			"GOAMD64": "v3",
		},
	}}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			env, err := buildEnv(test.platform, test.userEnv, test.configEnv)
			if err != nil {
				t.Fatalf("unexpected error running buildEnv(): %v", err)
			}
			envs := map[string]string{}
			for _, e := range env {
				split := strings.SplitN(e, "=", 2)
				envs[split[0]] = split[1]
			}
			for key, val := range test.expectedEnvs {
				if envs[key] != val {
					t.Errorf("buildEnv(): expected %s=%s, got %s=%s", key, val, key, envs[key])
				}
			}
		})
	}
}
