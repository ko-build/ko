package licensedb

import (
	"errors"
	paths "path"

	"github.com/go-enry/go-license-detector/v4/licensedb/api"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/go-enry/go-license-detector/v4/licensedb/internal"
)

var (
	// ErrNoLicenseFound is raised if no license files were found.
	ErrNoLicenseFound = errors.New("no license file was found")
)

// Detect returns the most probable reference licenses matched for the given
// file tree. Each match has the confidence assigned, from 0 to 1, 1 means 100% confident.
func Detect(fs filer.Filer) (map[string]api.Match, error) {
	files, err := fs.ReadDir("")
	if err != nil {
		return nil, err
	}
	fileNames := []string{}
	for _, file := range files {
		if !file.IsDir {
			fileNames = append(fileNames, file.Name)
		} else if internal.IsLicenseDirectory(file.Name) {
			// "license" directory, let's look inside
			subfiles, err := fs.ReadDir(file.Name)
			if err == nil {
				for _, subfile := range subfiles {
					if !subfile.IsDir {
						fileNames = append(fileNames, paths.Join(file.Name, subfile.Name))
					}
				}
			}
		}
	}
	candidates := internal.ExtractLicenseFiles(fileNames, fs)
	licenses := internal.InvestigateLicenseTexts(candidates)
	if len(licenses) > 0 {
		return licenses, nil
	}
	// Plan B: take the README, find the section about the license and apply NER
	candidates = internal.ExtractReadmeFiles(fileNames, fs)
	if len(candidates) == 0 {
		return nil, ErrNoLicenseFound
	}
	licenses = internal.InvestigateReadmeTexts(candidates, fs)
	if len(licenses) == 0 {
		return nil, ErrNoLicenseFound
	}
	return licenses, nil
}

// Preload database with licenses - load internal database from assets into memory.
// This method is an optimization for cases when the `Detect` method should return fast,
// e.g. in HTTP web servers where connection timeout can occur during detect
// `Preload` method could be called before server startup.
// This method os optional and it's not required to be called, other APIs loads license database
// lazily on first invocation.
func Preload() {
	internal.Preload()
}
