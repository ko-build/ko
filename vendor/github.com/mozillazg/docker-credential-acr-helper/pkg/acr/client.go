package acr

import (
	"time"
)

type Client struct{}

type Credentials struct {
	UserName   string
	Password   string
	ExpireTime time.Time
}

func (c *Client) GetCredentials(serverURL string) (*Credentials, error) {
	registry, err := parseServerURL(serverURL)
	if err != nil {
		return nil, err
	}

	if registry.IsEE {
		client, err := newEEClient(registry.Region)
		if err != nil {
			return nil, err
		}
		if registry.InstanceId == "" {
			instanceId, err := client.getInstanceId(registry.InstanceName)
			if err != nil {
				return nil, err
			}
			registry.InstanceId = instanceId
		}
		return client.getCredentials(registry.InstanceId)
	}

	client, err := newPersonClient(registry.Region)
	if err != nil {
		return nil, err
	}
	return client.getCredentials()
}
