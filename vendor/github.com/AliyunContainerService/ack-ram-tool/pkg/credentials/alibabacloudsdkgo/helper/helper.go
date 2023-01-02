package helper

import (
	"fmt"
	"github.com/aliyun/credentials-go/credentials"
	"os"
)

const (
	EnvRoleArn         = "ALIBABA_CLOUD_ROLE_ARN"
	EnvOidcProviderArn = "ALIBABA_CLOUD_OIDC_PROVIDER_ARN"
	EnvOidcTokenFile   = "ALIBABA_CLOUD_OIDC_TOKEN_FILE"
)

func HaveOidcCredentialRequiredEnv() bool {
	return os.Getenv(EnvRoleArn) != "" &&
		os.Getenv(EnvOidcProviderArn) != "" &&
		os.Getenv(EnvOidcTokenFile) != ""
}

func NewOidcCredential(sessionName string) (credential credentials.Credential, err error) {
	return GetOidcCredential(sessionName)
}

// Deprecated: Use NewOidcCredential instead
func GetOidcCredential(sessionName string) (credential credentials.Credential, err error) {
	roleArn := os.Getenv(EnvRoleArn)
	oidcArn := os.Getenv(EnvOidcProviderArn)
	tokenFile := os.Getenv(EnvOidcTokenFile)
	if roleArn == "" {
		return nil, fmt.Errorf("environment variable %q is missing", EnvRoleArn)
	}
	if oidcArn == "" {
		return nil, fmt.Errorf("environment variable %q is missing", EnvOidcProviderArn)
	}
	if tokenFile == "" {
		return nil, fmt.Errorf("environment variable %q is missing", EnvOidcTokenFile)
	}
	if _, err := os.Stat(tokenFile); err != nil {
		return nil, fmt.Errorf("unable to read file at %q: %s", tokenFile, err)
	}

	config := new(credentials.Config).
		SetType("oidc_role_arn").
		SetOIDCProviderArn(oidcArn).
		SetOIDCTokenFilePath(tokenFile).
		SetRoleArn(roleArn).
		SetRoleSessionName(sessionName)

	return credentials.NewCredential(config)
}
