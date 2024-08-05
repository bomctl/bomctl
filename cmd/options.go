// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/options.go
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
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	directoryValue      string
	existingFileValue   string
	outputFileValue     string
	urlValue            string
	directorySliceValue []string
	fileSliceValue      []string
	urlSliceValue       []string
)

var (
	errDirNotFound  = errors.New("not a directory or does not exist")
	errFileNotFound = errors.New("not a file or does not exist")
)

func checkDirectory(value string) {
	directory, err := os.Stat(value)

	if errors.Is(err, fs.ErrNotExist) || !directory.Mode().IsDir() {
		cobra.CheckErr(fmt.Errorf("%w: %s", errDirNotFound, value))
	}
}

func checkFile(value string) {
	file, err := os.Stat(value)

	if errors.Is(err, fs.ErrNotExist) || !file.Mode().IsRegular() {
		cobra.CheckErr(fmt.Errorf("%w: %s", errFileNotFound, value))
	}
}

func (dv *directoryValue) String() string       { return fmt.Sprintf("%v", *dv) }
func (dsv *directorySliceValue) String() string { return fmt.Sprintf("%v", *dsv) }
func (efv *existingFileValue) String() string   { return fmt.Sprintf("%v", *efv) }
func (fsv *fileSliceValue) String() string      { return fmt.Sprintf("%v", *fsv) }
func (ofv *outputFileValue) String() string     { return fmt.Sprintf("%v", *ofv) }

func (uv *urlValue) String() string       { return fmt.Sprintf("%v", *uv) }
func (usv *urlSliceValue) String() string { return fmt.Sprintf("%v", *usv) }

func (dv *directoryValue) Set(value string) error {
	checkDirectory(value)
	*dv = directoryValue(value)

	return nil
}

func (dsv *directorySliceValue) Set(value string) error {
	checkDirectory(value)
	*dsv = append(*dsv, value)

	return nil
}

func (efv *existingFileValue) Set(value string) error {
	checkFile(value)
	*efv = existingFileValue(value)

	return nil
}

func (fsv *fileSliceValue) Set(value string) error {
	checkFile(value)
	*fsv = append(*fsv, value)

	return nil
}

func (ofv *outputFileValue) Set(value string) error {
	*ofv = outputFileValue(value)

	return nil
}

func (uv *urlValue) Set(value string) error {
	*uv = urlValue(value)

	return nil
}

func (usv *urlSliceValue) Set(value string) error {
	*usv = append(*usv, value)

	return nil
}

const (
	valueTypeDir    string = "DIRECTORY"
	valueTypeFile   string = "FILE"
	valueTypeURL    string = "URL"
	valueTypeString string = "STRING"
)

func (*directoryValue) Type() string      { return valueTypeDir }
func (*directorySliceValue) Type() string { return valueTypeDir }
func (*existingFileValue) Type() string   { return valueTypeFile }
func (*fileSliceValue) Type() string      { return valueTypeFile }
func (*outputFileValue) Type() string     { return valueTypeFile }
func (*urlValue) Type() string            { return valueTypeURL }
func (*urlSliceValue) Type() string       { return valueTypeURL }

var (
	_ pflag.Value = (*directoryValue)(nil)
	_ pflag.Value = (*directorySliceValue)(nil)
	_ pflag.Value = (*existingFileValue)(nil)
	_ pflag.Value = (*fileSliceValue)(nil)
	_ pflag.Value = (*outputFileValue)(nil)
	_ pflag.Value = (*urlValue)(nil)
	_ pflag.Value = (*urlSliceValue)(nil)
)
