package repository

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/MakeNowJust/heredoc"
)

const apkIndexFilename = "APKINDEX"
const descriptionFilename = "DESCRIPTION"

// Go template for generating the APKINDEX file from an ApkIndex struct
var apkIndexTemplate = template.Must(template.New(apkIndexFilename).Funcs(
	template.FuncMap{
		// Helper function to join slice of string by space
		"join": func(s []string) string {
			return strings.Join(s, " ")
		},
	}).Parse(heredoc.Doc(`C:{{.ChecksumString}}
		P:{{.Name}}
		V:{{.Version}}
		{{- if .Arch}}
		A:{{.Arch}}
		{{- end }}
		{{- if .Size}}
		S:{{.Size}}
		{{- end }}
		{{- if .InstalledSize}}
		I:{{.InstalledSize}}
		{{- end}}
		T:{{.Description}}
		{{- if .URL}}
		U:{{.URL}}
		{{- end}}
		{{- if .License}}
		L:{{.License}}
		{{- end}}
		{{- if .Origin}}
		o:{{.Origin}}
		{{- end}}
		{{- if .Maintainer}}
		m:{{.Maintainer}}
		{{- end}}
		{{- if and .BuildTime (not .BuildTime.IsZero)}}
		t:{{.BuildTime.Unix}}
		{{- end}}
		{{- if .RepoCommit}}
		c:{{.RepoCommit}}
		{{- end}}
		{{- if .Dependencies}}
		D:{{join .Dependencies}}
		{{- end}}
		{{- if .InstallIf}}
		i:{{.InstallIf}}
		{{- end}}
		{{- if .Provides}}
		p:{{join .Provides}}
		{{- end}}
		{{- if .Replaces}}
		r:{{.Replaces}}
		{{- end}}
		{{- if .ProviderPriority}}
		k:{{.ProviderPriority}}
		{{- end}}

	`)))

type ApkIndex struct {
	Signature   []byte
	Description string
	Packages    []*Package
}

// ParsePackageIndex parses a plain (uncompressed) APKINDEX file. It returns an
// ApkIndex struct
func ParsePackageIndex(apkIndexUnpacked io.Reader) (packages []*Package, err error) {
	if closer, ok := apkIndexUnpacked.(io.Closer); ok {
		defer closer.Close()
	}

	indexScanner := bufio.NewScanner(apkIndexUnpacked)

	pkg := &Package{}
	linenr := 1

	for indexScanner.Scan() {
		line := indexScanner.Text()
		if len(line) == 0 {
			if pkg.Name != "" {
				packages = append(packages, pkg)
			}
			pkg = &Package{}
			continue
		}

		if len(line) > 1 && line[1:2] != ":" {
			return nil, fmt.Errorf("cannot parse line %d: expected \":\" in not found", linenr)
		}

		token := line[:1]
		val := line[2:]

		switch token {
		case "P":
			pkg.Name = val
		case "V":
			pkg.Version = val
		case "A":
			pkg.Arch = val
		case "L":
			pkg.License = val
		case "T":
			pkg.Description = val
		case "o":
			pkg.Origin = val
		case "m":
			pkg.Maintainer = val
		case "U":
			pkg.URL = val
		case "D":
			pkg.Dependencies = strings.Split(val, " ")
		case "p":
			pkg.Provides = strings.Split(val, " ")
		case "c":
			pkg.RepoCommit = val
		case "t":
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse build time %s: %w", val, err)
			}
			pkg.BuildTime = time.Unix(i, 0)
		case "i":
			pkg.InstallIf = strings.Split(val, " ")
		case "S":
			size, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse size field %s: %w", val, err)
			}
			pkg.Size = size
		case "I":
			installedSize, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse installed size field %s: %w", val, err)
			}
			pkg.InstalledSize = installedSize
		case "k":
			priority, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse provider priority field %s: %w", val, err)
			}
			pkg.ProviderPriority = priority
		case "C":
			// Handle SHA1 checksums:
			if strings.HasPrefix(val, "Q1") {
				checksum, err := base64.StdEncoding.DecodeString(val[2:])
				if err != nil {
					return nil, err
				}
				pkg.Checksum = checksum
			}
		}

		linenr++
	}

	return
}

func IndexFromArchive(archive io.ReadCloser) (apkindex *ApkIndex, err error) {
	gzipReader, err := gzip.NewReader(archive)

	if err != nil {
		return
	}

	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	apkindex = &ApkIndex{}

	for {
		hdr, tarErr := tarReader.Next()

		if tarErr == io.EOF {
			break
		}

		if tarErr != nil {
			return nil, tarErr
		}

		switch hdr.Name {
		case apkIndexFilename:
			apkindex.Packages, err = ParsePackageIndex(io.NopCloser(tarReader))
			if err != nil {
				return
			}
		case descriptionFilename:
			description, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			apkindex.Description = string(description)
		default:
			if strings.HasPrefix(hdr.Name, ".SIGN.") {
				apkindex.Signature, err = io.ReadAll(tarReader)
			} else {
				return nil, fmt.Errorf("unexpected file found in APKINDEX: %s", hdr.Name)
			}
		}
	}

	return apkindex, nil
}

func ArchiveFromIndex(apkindex *ApkIndex) (archive io.Reader, err error) {
	// Execute the template and append output for each package in the index
	var apkindexContents bytes.Buffer
	for _, pkg := range apkindex.Packages {
		if len(pkg.Name) == 0 {
			continue
		}
		err = apkIndexTemplate.Execute(&apkindexContents, pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template for package %s: %w", pkg.Name, err)
		}
	}

	// Create the tarball
	var tarballContents bytes.Buffer
	gw := gzip.NewWriter(&tarballContents)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add APKINDEX and DESCRIPTION files to the tarball
	for _, item := range []struct {
		filename string
		contents []byte
	}{
		{apkIndexFilename, apkindexContents.Bytes()},
		{descriptionFilename, []byte(apkindex.Description)},
	} {
		var info os.FileInfo = &tarballItemFileInfo{item.filename, int64(len(item.contents))}
		header, err := tar.FileInfoHeader(info, item.filename)
		if err != nil {
			return nil, fmt.Errorf("creating tar header for %s: %w", item.filename, err)
		}
		header.Name = item.filename
		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("writing tar header for %s: %w", item.filename, err)
		}
		if _, err = io.Copy(tw, bytes.NewReader(item.contents)); err != nil {
			return nil, fmt.Errorf("copying tar contents for %s: %w", item.filename, err)
		}
	}

	// Return io.ReadCloser representing the tarball
	return &tarballContents, nil
}

// This type implements os.FileInfo, allowing us to construct
// a tar header without needing to run os.Stat on a file
type tarballItemFileInfo struct {
	name string
	size int64
}

func (info *tarballItemFileInfo) Name() string       { return info.name }
func (info *tarballItemFileInfo) Size() int64        { return info.size }
func (info *tarballItemFileInfo) Mode() os.FileMode  { return 0644 }
func (info *tarballItemFileInfo) ModTime() time.Time { return time.Time{} }
func (info *tarballItemFileInfo) IsDir() bool        { return false }
func (info *tarballItemFileInfo) Sys() interface{}   { return nil }

var _ os.FileInfo = (*tarballItemFileInfo)(nil)
