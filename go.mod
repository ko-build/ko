module github.com/google/ko

go 1.16

require (
	github.com/aws/aws-sdk-go-v2/config v1.15.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecr v1.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.13.0 // indirect
	github.com/awslabs/amazon-ecr-credential-helper/ecr-login v0.0.0-20220228164355-396b2034c795
	github.com/chrismellard/docker-credential-acr-env v0.0.0-20220119192733-fe33c00cee21
	github.com/containerd/stargz-snapshotter/estargz v0.11.3
	github.com/docker/docker v20.10.14+incompatible
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/go-training/helloworld v0.0.0-20200225145412-ba5f4379d78b
	github.com/google/go-cmp v0.5.7
	github.com/google/go-containerregistry v0.8.1-0.20220209165246-a44adc326839
	github.com/opencontainers/image-spec v1.0.3-0.20220114050600-8b9d41f48198
	github.com/sigstore/cosign v1.3.2-0.20211120003522-90e2dcfe7b92
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.11.0
	go.uber.org/automaxprocs v1.4.1-0.20220314153950-975e177ad84f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.10
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/apimachinery v0.23.5
	sigs.k8s.io/kind v0.12.0
)
