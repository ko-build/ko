// +build go1.18

package sbom

import (
	"debug/buildinfo"
	"runtime/debug"
)

type BuildInfo buildinfo.BuildInfo

func (bi *BuildInfo) UnmarshalText(data []byte) error {
	dbi := (*debug.BuildInfo)(bi)
	if err := dbi.UnmarshalText(data); err != nil {
		return err
	}
	bi = (*BuildInfo)(dbi)
	return nil
}
