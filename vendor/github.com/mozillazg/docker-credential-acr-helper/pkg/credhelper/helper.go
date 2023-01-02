package credhelper

import (
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/mozillazg/docker-credential-acr-helper/pkg/acr"
	"github.com/mozillazg/docker-credential-acr-helper/pkg/version"
	"github.com/sirupsen/logrus"
)

var errNotImplemented = errors.New("not implemented")

type ACRHelper struct {
	client *acr.Client
	logger *logrus.Logger
}

func NewACRHelper() *ACRHelper {
	return &ACRHelper{
		client: &acr.Client{},
		logger: logrus.StandardLogger(),
	}
}

func (a *ACRHelper) WithLoggerOut(w io.Writer) *ACRHelper {
	logger := logrus.New()
	logger.Out = w
	a.logger = logger
	return a
}

func (a *ACRHelper) Get(serverURL string) (string, string, error) {
	// TODO: add cache
	cred, err := a.client.GetCredentials(serverURL)
	if err != nil {
		a.logger.WithField("name", version.ProjectName).
			WithField("serverURL", serverURL).
			WithError(err).Error("get credentials failed")
		return "", "", fmt.Errorf("%s: get credentials for %q failed: %s",
			version.ProjectName, serverURL, err)
	}
	return cred.UserName, cred.Password, nil
}

func (a *ACRHelper) Add(creds *credentials.Credentials) error {
	return errNotImplemented
}

func (a *ACRHelper) Delete(serverURL string) error {
	return errNotImplemented
}

func (a *ACRHelper) List() (map[string]string, error) {
	return nil, errNotImplemented
}
