package repository

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/ini.v1"
)

// Package represents a single package with the information present in an
// APKINDEX.
type Package struct {
	Name             string
	Version          string
	Arch             string
	Description      string
	License          string
	Origin           string
	Maintainer       string
	URL              string
	Checksum         []byte
	Dependencies     []string
	Provides         []string
	InstallIf        []string
	Size             uint64
	InstalledSize    uint64
	ProviderPriority uint64
	BuildTime        time.Time
	RepoCommit       string
	Replaces         string
}

// Returns the package filename as it's named in a repository.
func (p *Package) Filename() string {
	return fmt.Sprintf("%s-%s.apk", p.Name, p.Version)
}

// ChecksumString returns a human-readable version of the checksum.
func (p *Package) ChecksumString() string {
	return "Q1" + base64.StdEncoding.EncodeToString(p.Checksum)
}

// ParsePackage parses a .apk file and returns a Package struct
func ParsePackage(apkPackage io.Reader) (*Package, error) {
	expanded, err := expandApk(apkPackage)
	if err != nil {
		return nil, fmt.Errorf("expandApk(): %v", err)
	}
	defer os.RemoveAll(expanded.TempDir)

	checksum, err := getApkChecksum(expanded.ControlDataTarGzFilename)
	if err != nil {
		return nil, fmt.Errorf("getApkChecksum(): %v", err)
	}

	file, err := os.Open(expanded.ControlDataTarGzFilename)
	if err != nil {
		return nil, fmt.Errorf("os.Open(): %v", err)
	}

	gzipRead, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader(): %v", err)
	}
	defer gzipRead.Close()

	tarRead := tar.NewReader(gzipRead)
	if _, err = tarRead.Next(); err != nil {
		return nil, fmt.Errorf("tarRead.Next(): %v", err)
	}

	cfg, err := ini.ShadowLoad(tarRead)
	if err != nil {
		return nil, fmt.Errorf("ini.ShadowLoad(): %w", err)
	}

	info := new(apkInfo)
	if err = cfg.MapTo(info); err != nil {
		return nil, fmt.Errorf("cfg.MapTo(): %w", err)
	}

	return &Package{
		Name:          info.PKGName,
		Version:       info.PKGVer,
		Arch:          info.Arch,
		Description:   info.PKGDesc,
		License:       info.License,
		Origin:        info.Origin,
		Maintainer:    info.Maintainer,
		URL:           info.URL,
		Checksum:      checksum,
		Dependencies:  info.Depend,
		Provides:      info.Provides,
		Size:          uint64(expanded.Size),
		InstalledSize: uint64(info.Size),
		RepoCommit:    info.Commit,
	}, nil
}

// TODO: remove this struct if possible and just use Package type
type apkInfo struct {
	PKGName    string   `ini:"pkgname"`
	PKGVer     string   `ini:"pkgver"`
	PKGDesc    string   `ini:"pkgdesc"`
	URL        string   `ini:"url"`
	Size       int      `ini:"size"`
	Arch       string   `ini:"arch"`
	Origin     string   `ini:"origin"`
	Commit     string   `ini:"commit"`
	Maintainer string   `ini:"maintainer"`
	License    string   `ini:"license"`
	Depend     []string `ini:"depend,,allowshadow"`
	Provides   []string `ini:"provides,,allowshadow"`
}
