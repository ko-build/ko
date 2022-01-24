package cmd

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type platformValue struct {
	platform *v1.Platform
}

func (pv *platformValue) Set(platform string) error {
	p, err := v1.PlatformFromString(platform)
	if err != nil {
		return err
	}
	pv.platform = p
	return nil
}

func (pv *platformValue) String() string {
	if pv == nil || pv.platform == nil {
		return "none"
	}
	return pv.platform.String()
}

func (pv *platformValue) Type() string {
	return "platform"
}
