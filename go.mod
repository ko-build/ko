module github.com/google/ko

go 1.14

require (
	github.com/docker/cli v0.0.0-20200303162255-7d407207c304 // indirect
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-logr/logr v0.2.1 // indirect
	github.com/go-training/helloworld v0.0.0-20200225145412-ba5f4379d78b
	github.com/google/go-cmp v0.4.1
	github.com/google/go-containerregistry v0.1.3
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mattmoor/dep-notify v0.0.0-20190205035814-a45dec370a17
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200924205911-8a9a89368bd3
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	gotest.tools/v3 v3.0.2 // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/code-generator v0.19.0-alpha.1 // indirect
	k8s.io/gengo v0.0.0-20200728071708-7794989d0000 // indirect
	k8s.io/klog/v2 v2.3.0 // indirect
	sigs.k8s.io/kind v0.8.1
	sigs.k8s.io/structured-merge-diff/v4 v4.0.1 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
)
