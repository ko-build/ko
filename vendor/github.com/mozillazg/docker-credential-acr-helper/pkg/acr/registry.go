package acr

import (
	"errors"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var errUnknownDomain = errors.New("unknown domain")
var domainPattern = regexp.MustCompile(
	`^(?:(?P<instanceName>[^.\s]+)-)?registry(?:-intl)?(?:-vpc)?(?:-internal)?(?:\.distributed)?\.(?P<region>[^.]+)\.(?:cr\.)?aliyuncs\.com`)

const (
	urlPrefix      = "https://"
	hostNameSuffix = ".aliyuncs.com"

	envInstanceId = "DOCKER_CREDENTIAL_ACR_HELPER_INSTANCE_ID"
	envRegion     = "DOCKER_CREDENTIAL_ACR_HELPER_REGION"
)

type Registry struct {
	IsEE         bool
	InstanceId   string
	InstanceName string
	Region       string
	Domain       string
}

func parseServerURL(rawURL string) (*Registry, error) {
	instanceId := os.Getenv(envInstanceId)
	if instanceId == "" {
		if !strings.Contains(rawURL, hostNameSuffix) {
			return nil, errUnknownDomain
		}
	}

	if !strings.HasPrefix(rawURL, urlPrefix) {
		rawURL = urlPrefix + rawURL
	}
	serverURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	domain := serverURL.Hostname()

	if instanceId == "" {
		if !strings.HasSuffix(domain, hostNameSuffix) {
			return nil, errUnknownDomain
		}
	}

	registry := &Registry{
		IsEE:         instanceId != "",
		InstanceId:   instanceId,
		InstanceName: "",
		Region:       os.Getenv(envRegion),
		Domain:       domain,
	}

	// parse domain to get acr ee instance info
	if registry.InstanceId == "" || registry.Region == "" {
		subItems := domainPattern.FindStringSubmatch(domain)
		if len(subItems) != 3 {
			return nil, errUnknownDomain
		}
		registry.InstanceName = subItems[1]
		registry.Region = subItems[2]
		registry.IsEE = registry.InstanceName != ""
	}

	return registry, nil
}
