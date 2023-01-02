package credentials

import (
	"errors"
)

type providerChain struct {
	Providers []Provider
}

var defaultproviders = []Provider{providerEnv, providerProfile, providerInstance}
var defaultChain = newProviderChain(defaultproviders)

func newProviderChain(providers []Provider) Provider {
	return &providerChain{
		Providers: providers,
	}
}

func (p *providerChain) resolve() (*Config, error) {
	for _, provider := range p.Providers {
		config, err := provider.resolve()
		if err != nil {
			return nil, err
		} else if config == nil {
			continue
		}
		return config, err
	}
	return nil, errors.New("No credential found")

}
