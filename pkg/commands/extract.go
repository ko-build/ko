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

package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dprotaso/go-yit"
	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// addExtract augments our CLI surface with extract.
func addExtract(topLevel *cobra.Command) {
	fo := &options.FilenameOptions{}
	extract := &cobra.Command{
		Use:   "extract -f FILENAME...",
		Short: "Extract ko-built image references from YAML configs",
		Long:  `This sub-command extracts image references detected to have been built by ko`,
		Example: `
# Extract and sign ko-built images:
ko extract -f release.yaml | xargs cosign sign ...

# Extract and rebase ko-built images:
ko extract -f release.yaml | xargs crane rebase ...
`,
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			dockerRepo := os.Getenv("KO_DOCKER_REPO")
			if dockerRepo == "" {
				return errors.New("KO_DOCKER_REPO environment variable is unset")
			}

			found := map[string]struct{}{}

			for f := range options.EnumerateFiles(fo) {
				var b []byte
				var err error
				if f == "-" {
					b, err = ioutil.ReadAll(os.Stdin)
				} else {
					b, err = ioutil.ReadFile(f)
				}
				if err != nil {
					return err
				}

				// The loop is to support multi-document yaml files.
				// This is handled by using a yaml.Decoder and reading objects until io.EOF, see:
				// https://godoc.org/gopkg.in/yaml.v3#Decoder.Decode
				decoder := yaml.NewDecoder(bytes.NewBuffer(b))
				var docNodes []*yaml.Node
				for {
					var doc yaml.Node
					if err := decoder.Decode(&doc); err != nil {
						if err == io.EOF {
							break
						}
						return err
					}
					docNodes = append(docNodes, &doc)
				}
				for _, doc := range docNodes {
					it := yit.FromNode(doc).
						RecurseNodes().
						Filter(yit.StringValue)
					for node, ok := it(); ok; node, ok = it() {
						ref := strings.TrimSpace(node.Value)
						if strings.HasPrefix(ref, dockerRepo) {
							if _, ok := found[ref]; !ok {
								found[ref] = struct{}{}
								fmt.Println(ref)
							}
						}
					}
				}
			}
			return nil
		},
	}
	options.AddFileArg(extract, fo)
	topLevel.AddCommand(extract)
}
