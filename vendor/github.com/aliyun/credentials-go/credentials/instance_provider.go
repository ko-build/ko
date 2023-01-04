package credentials

import (
	"os"

	"github.com/alibabacloud-go/tea/tea"
)

type instanceCredentialsProvider struct{}

var providerInstance = new(instanceCredentialsProvider)

func newInstanceCredentialsProvider() Provider {
	return &instanceCredentialsProvider{}
}

func (p *instanceCredentialsProvider) resolve() (*Config, error) {
	roleName, ok := os.LookupEnv(ENVEcsMetadata)
	if !ok {
		return nil, nil
	}

	config := &Config{
		Type:     tea.String("ecs_ram_role"),
		RoleName: tea.String(roleName),
	}
	return config, nil
}
