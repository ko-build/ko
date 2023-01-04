package acr

import (
	"os"
	"path/filepath"

	"github.com/AliyunContainerService/ack-ram-tool/pkg/credentials/alibabacloudsdkgo/helper"
	"github.com/aliyun/credentials-go/credentials"
	"github.com/mozillazg/docker-credential-acr-helper/pkg/version"
)

var defaultProfilePath = filepath.Join("~", ".alibabacloud", "credentials")

func getOpenapiAuth() (credentials.Credential, error) {
	profilePath := defaultProfilePath
	if os.Getenv(credentials.ENVCredentialFile) != "" {
		profilePath = os.Getenv(credentials.ENVCredentialFile)
	}
	path, err := expandPath(profilePath)
	if err == nil {
		if _, err := os.Stat(path); err == nil {
			_ = os.Setenv(credentials.ENVCredentialFile, path)
		}
	}
	var conf *credentials.Config

	if helper.HaveOidcCredentialRequiredEnv() {
		return helper.NewOidcCredential(version.ProjectName)
	}

	cred, err := credentials.NewCredential(conf)
	return cred, err
}

func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}
