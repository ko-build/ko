#!/usr/bin/env bash

# Copyright 2021 ko Build Authors All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

pushd "${PROJECT_ROOT}"
trap popd EXIT

# Verify that all source files are correctly formatted.
find . -name "*.go" | grep -v vendor/ | xargs gofmt -d -e -l

# Verify that generated Markdown docs are up-to-date.
tmpdir=$(mktemp -d)
go run cmd/help/main.go --dir "$tmpdir"
diff -Naur -I '###### Auto generated' "$tmpdir" docs/reference/
