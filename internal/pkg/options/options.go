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
package options //nolint:revive

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/formats"

	"github.com/bomctl/bomctl/internal/pkg/logger"
)

type (
	Options struct {
		Logger     *log.Logger
		ctx        context.Context
		CacheDir   string
		ConfigFile string
		Verbosity  int
	}

	Option func(*Options)

	ExportOptions struct {
		*Options
		OutputFile *os.File
		Format     formats.Format
	}

	FetchOptions struct {
		OutputFile *os.File
		*Options
		UseNetRC bool
	}

	ImportOptions struct {
		*Options
		InputFiles []*os.File
	}

	PushOptions struct {
		*Options
		Format   formats.Format
		UseTree  bool
		UseNetRC bool
	}
)

func New(opts ...Option) *Options {
	options := &Options{
		// Instantiates with a default unprefixed logger.
		Logger: logger.New(""),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func (o *Options) Context() context.Context {
	return o.ctx
}

func (o *Options) WithCacheDir(dir string) *Options {
	o.CacheDir = dir

	return o
}

func (o *Options) WithConfigFile(file string) *Options {
	o.ConfigFile = file

	return o
}

func (o *Options) WithContext(ctx context.Context) *Options {
	o.ctx = ctx

	return o
}

func (o *Options) WithLogger(l *log.Logger) *Options {
	o.Logger = l

	return o
}

func (o *Options) WithVerbosity(level int) *Options {
	o.Verbosity = level

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

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.WithContext(ctx)
	}
}

func WithLogger(l *log.Logger) Option {
	return func(o *Options) {
		o.WithLogger(l)
	}
}

func WithVerbosity(level int) Option {
	return func(o *Options) {
		o.WithVerbosity(level)
	}
}
