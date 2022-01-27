package internal

import (
	"bytes"
	"fmt"
	paths "path"
	"regexp"
	"strings"
	"sync"

	"github.com/go-enry/go-license-detector/v4/licensedb/api"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/go-enry/go-license-detector/v4/licensedb/internal/processors"
)

var (
	globalLicenseDB struct {
		sync.Once
		*database
	}
	globalLicenseDatabase = func() *database {
		globalLicenseDB.Once.Do(func() {
			globalLicenseDB.database = loadLicenses()
		})
		return globalLicenseDB.database
	}

	// Base names of guessable license files
	licenseFileNames = []string{
		"li[cs]en[cs]e(s?)",
		"legal",
		"copy(left|right|ing)",
		"unlicense",
		"l?gpl([-_ v]?)(\\d\\.?\\d)?",
		"bsd",
		"mit",
		"apache",
	}

	// License file extensions. Combined with the fileNames slice
	// to create a set of files we can reasonably assume contain
	// licensing information.
	fileExtensions = []string{
		"",
		".md",
		".rst",
		".html",
		".txt",
	}

	filePreprocessors = map[string]func([]byte) []byte{
		".md":   processors.Markdown,
		".rst":  processors.RestructuredText,
		".html": processors.HTML,
	}

	licenseFileRe = regexp.MustCompile(
		fmt.Sprintf("^(|.*[-_. ])(%s)(|[-_. ].*)$",
			strings.Join(licenseFileNames, "|")))

	readmeFileRe = regexp.MustCompile(fmt.Sprintf("^(readme|guidelines)(%s)$",
		strings.Replace(strings.Join(fileExtensions, "|"), ".", "\\.", -1)))

	licenseDirectoryRe = regexp.MustCompile(fmt.Sprintf(
		"^(%s)$", strings.Join(licenseFileNames, "|")))
)

func investigateCandidates(candidates map[string][]byte, f func(text []byte) map[string]float32) map[string]api.Match {
	matches := make(map[string]api.Match)
	for file, text := range candidates {
		candidates := f(text)
		for name, sim := range candidates {
			match := matches[name]
			if match.Files == nil {
				match.Files = make(map[string]float32)
			}
			match.Files[file] = sim
			if sim > match.Confidence {
				match.Confidence = sim
				match.File = file
			}
			matches[name] = match
		}
	}
	return matches
}

// ExtractLicenseFiles returns the list of possible license texts.
// The file names are matched against the template.
// Reader is used to to read file contents.
func ExtractLicenseFiles(files []string, fs filer.Filer) map[string][]byte {
	candidates := make(map[string][]byte)
	for _, file := range files {
		if licenseFileRe.MatchString(strings.ToLower(paths.Base(file))) {
			text, err := fs.ReadFile(file)
			if len(text) < 128 {
				// e.g. https://github.com/Unitech/pm2/blob/master/LICENSE
				realText, err := fs.ReadFile(string(bytes.TrimSpace(text)))
				if err == nil {
					file = string(bytes.TrimSpace(text))
					text = realText
				}
			}
			if err == nil {
				if preprocessor, exists := filePreprocessors[paths.Ext(file)]; exists {
					text = preprocessor(text)
				}
				candidates[file] = text
			}
		}
	}
	return candidates
}

// InvestigateLicenseTexts takes the list of candidate license texts and returns the most probable
// reference licenses matched. Each match has the confidence assigned, from 0 to 1, 1 means 100% confident.
// Furthermore, each match contains a mapping of filename to the confidence that file produced.
func InvestigateLicenseTexts(candidates map[string][]byte) map[string]api.Match {
	return investigateCandidates(candidates, InvestigateLicenseText)
}

// InvestigateLicenseText takes the license text and returns the most probable reference licenses matched.
// Each match has the confidence assigned, from 0 to 1, 1 means 100% confident.
func InvestigateLicenseText(text []byte) map[string]float32 {
	return globalLicenseDatabase().QueryLicenseText(string(text))
}

// ExtractReadmeFiles searches for README files.
// Reader is used to to read file contents.
func ExtractReadmeFiles(files []string, fs filer.Filer) map[string][]byte {
	candidates := make(map[string][]byte)
	for _, file := range files {
		if readmeFileRe.MatchString(strings.ToLower(file)) {
			text, err := fs.ReadFile(file)
			if err == nil {
				if preprocessor, exists := filePreprocessors[paths.Ext(file)]; exists {
					text = preprocessor(text)
				}
				candidates[file] = text
			}
		}
	}
	return candidates
}

// InvestigateReadmeTexts scans README files for licensing information and outputs the
// probable names using NER.
func InvestigateReadmeTexts(candidtes map[string][]byte, fs filer.Filer) map[string]api.Match {
	return investigateCandidates(candidtes, func(text []byte) map[string]float32 {
		return InvestigateReadmeText(text, fs)
	})
}

// InvestigateReadmeText scans the README file for licensing information and outputs probable
// names found with Named Entity Recognition from NLP.
func InvestigateReadmeText(text []byte, fs filer.Filer) map[string]float32 {
	return globalLicenseDatabase().QueryReadmeText(string(text), fs)
}

// IsLicenseDirectory indicates whether the directory is likely to contain licenses.
func IsLicenseDirectory(fileName string) bool {
	return licenseDirectoryRe.MatchString(strings.ToLower(fileName))
}

// Preload license database
func Preload() {
	_ = globalLicenseDatabase()
}
