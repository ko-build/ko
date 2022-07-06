package platform

import (
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Matcher struct {
	Spec      []string
	Platforms []v1.Platform
}

func IsWasi(p *v1.Platform) bool {
	return p.OS == Wasi.OS && p.Architecture == Wasi.Architecture
}

var Wasi = &v1.Platform{
	OS:           "wasm",
	Architecture: "wasi",
}

func ParseSpec(spec []string) (*Matcher, error) {
	// Don't bother parsing "all".
	// Empty slice should never happen because we default to linux/amd64 (or GOOS/GOARCH).
	if len(spec) == 0 || spec[0] == "all" {
		return &Matcher{Spec: spec}, nil
	}

	platforms := []v1.Platform{}
	for _, s := range spec {
		p, err := v1.ParsePlatform(s)
		if err != nil {
			return nil, err
		}
		platforms = append(platforms, *p)
	}
	return &Matcher{Spec: spec, Platforms: platforms}, nil
}

func (m *Matcher) Matches(base *v1.Platform) bool {
	if len(m.Spec) > 0 && m.Spec[0] == "all" {
		return true
	}

	// Don't build anything without a platform field unless "all". Unclear what we should do here.
	if base == nil {
		return false
	}

	for _, p := range m.Platforms {
		if p.OS != "" && base.OS != p.OS {
			continue
		}
		if p.Architecture != "" && base.Architecture != p.Architecture {
			continue
		}
		if p.Variant != "" && base.Variant != p.Variant {
			continue
		}

		// Windows is... weird. Windows base images use osversion to
		// communicate what Windows version is used, which matters for image
		// selection at runtime.
		//
		// Windows osversions include the usual major/minor/patch version
		// components, as well as an incrementing "build number" which can
		// change when new Windows base images are released.
		//
		// In order to avoid having to match the entire osversion including the
		// incrementing build number component, we allow matching a platform
		// that only matches the first three osversion components, only for
		// Windows images.
		//
		// If the X.Y.Z components don't match (or aren't formed as we expect),
		// the platform doesn't match. Only if X.Y.Z matches and the extra
		// build number component doesn't, do we consider the platform to
		// match.
		//
		// Ref: https://docs.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/version-compatibility?tabs=windows-server-2022%2Cwindows-10-21H1#build-number-new-release-of-windows
		if p.OSVersion != "" && p.OSVersion != base.OSVersion {
			if p.OS != "windows" {
				// osversion mismatch is only possibly allowed when os == windows.
				continue
			} else {
				if pcount, bcount := strings.Count(p.OSVersion, "."), strings.Count(base.OSVersion, "."); pcount == 2 && bcount == 3 {
					if p.OSVersion != base.OSVersion[:strings.LastIndex(base.OSVersion, ".")] {
						// If requested osversion is X.Y.Z and potential match is X.Y.Z.A, all of X.Y.Z must match.
						// Any other form of these osversions are not a match.
						continue
					}
				} else {
					// Partial osversion matching only allows X.Y.Z to match X.Y.Z.A.
					continue
				}
			}
		}
		return true
	}

	return false
}
