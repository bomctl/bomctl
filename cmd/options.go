// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © bomctl a Series of LF Projects, LLC
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
)

type (
	DirectoryValue      string
	ExistingFileValue   string
	OutputFileValue     string
	FormatStringValue   string
	FormatEncodingValue string
	URLValue            string
	DirectorySliceValue []string
	FileSliceValue      []string
	URLSliceValue       []string
	SBOMIDSliceValue    []string
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

func (dv *DirectoryValue) String() string       { return fmt.Sprintf("%v", *dv) }
func (dsv *DirectorySliceValue) String() string { return fmt.Sprintf("%v", *dsv) }
func (efv *ExistingFileValue) String() string   { return fmt.Sprintf("%v", *efv) }
func (fsv *FileSliceValue) String() string      { return fmt.Sprintf("%v", *fsv) }
func (ofv *OutputFileValue) String() string     { return fmt.Sprintf("%v", *ofv) }
func (fstv *FormatStringValue) String() string  { return fmt.Sprintf("%v", *fstv) }
func (fev *FormatEncodingValue) String() string { return fmt.Sprintf("%v", *fev) }

func (uv *URLValue) String() string       { return fmt.Sprintf("%v", *uv) }
func (usv *URLSliceValue) String() string { return fmt.Sprintf("%v", *usv) }

func (dv *DirectoryValue) Set(value string) error {
	checkDirectory(value)
	*dv = DirectoryValue(value)

	return nil
}

func (dsv *DirectorySliceValue) Set(value string) error {
	checkDirectory(value)
	*dsv = append(*dsv, value)

	return nil
}

func (efv *ExistingFileValue) Set(value string) error {
	checkFile(value)
	*efv = ExistingFileValue(value)

	return nil
}

func (fsv *FileSliceValue) Set(value string) error {
	checkFile(value)
	*fsv = append(*fsv, value)

	return nil
}

func (ofv *OutputFileValue) Set(value string) error {
	*ofv = OutputFileValue(value)

	return nil
}

func (fstv *FormatStringValue) Set(value string) error {
	*fstv = FormatStringValue(value)

	return nil
}

func (fev *FormatEncodingValue) Set(value string) error {
	*fev = FormatEncodingValue(value)

	return nil
}

func (uv *URLValue) Set(value string) error {
	*uv = URLValue(value)

	return nil
}

func (usv *URLSliceValue) Set(value string) error {
	*usv = append(*usv, value)

	return nil
}

const (
	valueTypeDir    string = "DIRECTORY"
	valueTypeFile   string = "FILE"
	valueTypeURL    string = "URL"
	valueTypeString string = "STRING"
)

func (dv *DirectoryValue) Type() string       { return valueTypeDir }
func (dsv *DirectorySliceValue) Type() string { return valueTypeDir }
func (efv *ExistingFileValue) Type() string   { return valueTypeFile }
func (fsv *FileSliceValue) Type() string      { return valueTypeFile }
func (ofv *OutputFileValue) Type() string     { return valueTypeFile }
func (fstv *FormatStringValue) Type() string  { return valueTypeString }
func (fev *FormatEncodingValue) Type() string { return valueTypeString }
func (uv *URLValue) Type() string             { return valueTypeURL }
func (usv *URLSliceValue) Type() string       { return valueTypeURL }
