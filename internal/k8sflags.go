// Copyright 2021 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"strings"

	"github.com/spf13/pflag"
)

// AddFlags adds kubectl global flags to the given flagset.
func AddFlags(f *KubectlFlags, flags *pflag.FlagSet) {
	flags.StringVar(&f.kubeConfig, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests. (DEPRECATED)")
	flags.StringVar(&f.cacheDir, "cache-dir", "", "Default cache directory (DEPRECATED)")
	flags.StringVar(&f.certFile, "client-certificate", "", "Path to a client certificate file for TLS (DEPRECATED)")
	flags.StringVar(&f.keyFile, "client-key", "", "Path to a client key file for TLS (DEPRECATED)")
	flags.StringVar(&f.bearerToken, "token", "", "Bearer token for authentication to the API server (DEPRECATED)")
	flags.StringVar(&f.impersonate, "as", "", "Username to impersonate for the operation (DEPRECATED)")
	flags.StringArrayVar(&f.impersonateGroup, "as-group", []string{}, "Group to impersonate for the operation, this flag can be repeated to specify multiple groups. (DEPRECATED)")
	flags.StringVar(&f.username, "username", "", "Username for basic authentication to the API server (DEPRECATED)")
	flags.StringVar(&f.password, "password", "", "Password for basic authentication to the API server (DEPRECATED)")
	flags.StringVar(&f.clusterName, "cluster", "", "The name of the kubeconfig cluster to use (DEPRECATED)")
	flags.StringVar(&f.authInfoName, "user", "", "The name of the kubeconfig user to use (DEPRECATED)")
	flags.StringVarP(&f.namespace, "namespace", "n", "", "If present, the namespace scope for this CLI request (DEPRECATED)")
	flags.StringVar(&f.context, "context", "", "The name of the kubeconfig context to use (DEPRECATED)")
	flags.StringVarP(&f.apiServer, "server", "s", "", "The address and port of the Kubernetes API server (DEPRECATED)")
	flags.StringVar(&f.tlsServerName, "tls-server-name", "", "Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used (DEPRECATED)")
	flags.BoolVar(&f.insecure, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure (DEPRECATED)")
	flags.StringVar(&f.caFile, "certificate-authority", "", "Path to a cert file for the certificate authority (DEPRECATED)")
	flags.StringVar(&f.timeout, "request-timeout", "", "The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (DEPRECATED)")
}

// KubectlFlags holds kubectl global flag values as parsed from flags.
type KubectlFlags struct {
	kubeConfig       string
	cacheDir         string
	certFile         string
	keyFile          string
	bearerToken      string
	impersonate      string
	impersonateGroup []string
	username         string
	password         string
	clusterName      string
	authInfoName     string
	context          string
	namespace        string
	apiServer        string
	tlsServerName    string
	insecure         bool
	caFile           string
	timeout          string
}

// Values returns a slice of flag values to pass to kubectl.
func (f KubectlFlags) Values() []string {
	var v []string
	if f.kubeConfig != "" {
		v = append(v, "--kubeconfig="+f.kubeConfig)
	}
	if f.cacheDir != "" {
		v = append(v, "--cache-dir="+f.cacheDir)
	}
	if f.certFile != "" {
		v = append(v, "--client-certificate="+f.certFile)
	}
	if f.keyFile != "" {
		v = append(v, "--client-key="+f.keyFile)
	}
	if f.bearerToken != "" {
		v = append(v, "--token="+f.bearerToken)
	}
	if f.impersonate != "" {
		v = append(v, "--as="+f.impersonate)
	}
	if len(f.impersonateGroup) > 0 {
		v = append(v, "--as-group="+strings.Join(f.impersonateGroup, ","))
	}
	if f.username != "" {
		v = append(v, "--username="+f.username)
	}
	if f.password != "" {
		v = append(v, "--password="+f.password)
	}
	if f.clusterName != "" {
		v = append(v, "--cluster="+f.clusterName)
	}
	if f.authInfoName != "" {
		v = append(v, "--user="+f.authInfoName)
	}
	if f.context != "" {
		v = append(v, "--context="+f.context)
	}
	if f.namespace != "" {
		v = append(v, "--namespace="+f.namespace)
	}
	if f.apiServer != "" {
		v = append(v, "--server="+f.apiServer)
	}
	if f.tlsServerName != "" {
		v = append(v, "--tls-server-name="+f.tlsServerName)
	}
	if f.insecure {
		v = append(v, "--insecure=true")
	}
	if f.caFile != "" {
		v = append(v, "--certificate-authority="+f.caFile)
	}
	if f.timeout != "" {
		v = append(v, "--request-timeout="+f.timeout)
	}
	return v
}
