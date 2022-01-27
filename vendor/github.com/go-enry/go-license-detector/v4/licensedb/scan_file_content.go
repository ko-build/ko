package licensedb

import (
	"github.com/go-enry/go-license-detector/v4/licensedb/internal"
)

// InvestigateLicenseText takes the license text and returns the most probable reference licenses matched.
// Each match has the confidence assigned, from 0 to 1, 1 means 100% confident.
func InvestigateLicenseText(text []byte) map[string]float32 {
	return internal.InvestigateLicenseText(text)
}
