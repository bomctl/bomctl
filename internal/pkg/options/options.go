// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/options/options.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// ------------------------------------------------------------------------
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ------------------------------------------------------------------------
package options

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/formats"
)

type (
	Options struct {
		Logger     *log.Logger
		CacheDir   string
		ConfigFile string
		Debug      bool
	}

	Option func(*Options)

	FetchOptions struct {
		OutputFile *os.File
		*Options
		UseNetRC bool
		Alias    string
		Tags     []string
	}

	PushOptions struct {
		*Options
		Format   formats.Format
		UseTree  bool
		UseNetRC bool
	}
)

func New(opts ...Option) *Options {
	options := &Options{}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func (o *Options) WithCacheDir(dir string) *Options {
	o.CacheDir = dir

	return o
}

func (o *Options) WithConfigFile(file string) *Options {
	o.ConfigFile = file

	return o
}

func (o *Options) WithDebug(debug bool) *Options {
	o.Debug = debug

	return o
}

func (o *Options) WithLogger(logger *log.Logger) *Options {
	o.Logger = logger

	return o
}

func WithCacheDir(dir string) Option {
	return func(o *Options) {
		o.WithCacheDir(dir)
	}
}

func WithConfigFile(file string) Option {
	return func(o *Options) {
		o.WithConfigFile(file)
	}
}

func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.WithDebug(debug)
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(o *Options) {
		o.WithLogger(logger)
	}
}
