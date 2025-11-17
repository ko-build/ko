// Copyright 2025 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package build

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
)

type Printer interface {
	Print(images map[string]name.Reference) error
}

type PrinterFunc func(images map[string]name.Reference) error

func (f PrinterFunc) Print(images map[string]name.Reference) error {
	return f(images)
}

func NewPrinter(w io.Writer) Printer {
	return PrinterFunc(Print)
}

func Print(images map[string]name.Reference) error {
	for _, img := range images {
		fmt.Println(img)
	}

	return nil
}

func NewJSONPrinter(w io.Writer, indent bool) Printer {
	encoder := json.NewEncoder(w)
	if indent {
		encoder.SetIndent("", "  ")
	}

	return &jsonPrinter{
		enc: encoder,
	}
}

type jsonPrinter struct {
	enc *json.Encoder
}

func (p *jsonPrinter) Print(images map[string]name.Reference) error {
	return p.enc.Encode(images)
}
