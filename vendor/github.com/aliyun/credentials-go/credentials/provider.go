package credentials

//Environmental virables that may be used by the provider
const (
	ENVCredentialFile  = "ALIBABA_CLOUD_CREDENTIALS_FILE"
	ENVEcsMetadata     = "ALIBABA_CLOUD_ECS_METADATA"
	PATHCredentialFile = "~/.alibabacloud/credentials"
)

// Provider will be implemented When you want to customize the provider.
type Provider interface {
	resolve() (*Config, error)
}
