// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/options.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------

package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

const (
	valueTypeChoice string = "CHOICE"
	valueTypeDir    string = "DIRECTORY"
	valueTypeFile   string = "FILE"
	valueTypeURL    string = "URL"
)

type (
	choiceValue struct {
		usage   string
		value   *string
		choices []string
	}

	directorySliceValue []string
	directoryValue      string
	existingFileValue   string
	fileSliceValue      []string
	outputFileValue     string
	urlSliceValue       []string
	urlValue            string
)

func (cv *choiceValue) Choices() string {
	return fmt.Sprintf("[%s]", strings.Join(cv.choices, ", "))
}

func (cv *choiceValue) CompletionFunc() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return cv.choices, cobra.ShellCompDirectiveDefault
	}
}

func (cv *choiceValue) Error() string {
	return "must be one of " + cv.Choices()
}

func (cv *choiceValue) Set(value string) error {
	var err error

	if *cv.value, err = sliceutil.Next(cv.choices, func(s string) bool { return s == value }); err != nil {
		return fmt.Errorf("%w", cv)
	}

	return nil
}

func (cv *choiceValue) String() string {
	return *cv.value
}

func (*choiceValue) Type() string {
	return valueTypeChoice
}

func (cv *choiceValue) Usage() string {
	return fmt.Sprintf("%s %s", cv.usage, cv.Choices())
}

func (dsv *directorySliceValue) Set(value string) error {
	checkDirectory(value)
	*dsv = append(*dsv, value)

	return nil
}

func (dsv *directorySliceValue) String() string {
	return fmt.Sprintf("%v", *dsv)
}

func (*directorySliceValue) Type() string {
	return valueTypeDir
}

func (dv *directoryValue) Set(value string) error {
	checkDirectory(value)
	*dv = directoryValue(value)

	return nil
}

func (dv *directoryValue) String() string {
	return fmt.Sprintf("%v", *dv)
}

func (*directoryValue) Type() string {
	return valueTypeDir
}

func (efv *existingFileValue) Set(value string) error {
	checkFile(value)
	*efv = existingFileValue(value)

	return nil
}

func (efv *existingFileValue) String() string {
	return fmt.Sprintf("%v", *efv)
}

func (*existingFileValue) Type() string {
	return valueTypeFile
}

func (fsv *fileSliceValue) Set(value string) error {
	checkFile(value)
	*fsv = append(*fsv, value)

	return nil
}

func (fsv *fileSliceValue) String() string {
	return fmt.Sprintf("%v", *fsv)
}

func (*fileSliceValue) Type() string {
	return valueTypeFile
}

func (ofv *outputFileValue) Set(value string) error {
	*ofv = outputFileValue(value)

	return nil
}

func (ofv *outputFileValue) String() string {
	return fmt.Sprintf("%v", *ofv)
}

func (*outputFileValue) Type() string {
	return valueTypeFile
}

func (usv *urlSliceValue) Set(value string) error {
	*usv = append(*usv, value)

	return nil
}

func (usv *urlSliceValue) String() string {
	return fmt.Sprintf("%v", *usv)
}

func (*urlSliceValue) Type() string {
	return valueTypeURL
}

func (uv *urlValue) Set(value string) error {
	*uv = urlValue(value)

	return nil
}

func (uv *urlValue) String() string {
	return fmt.Sprintf("%v", *uv)
}

func (*urlValue) Type() string {
	return valueTypeURL
}

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

func newChoiceValue(usage, value string, choices ...string) *choiceValue {
	return &choiceValue{
		usage:   usage,
		value:   &value,
		choices: append([]string{value}, choices...),
	}
}

var _ = []pflag.Value{
	(*choiceValue)(nil),
	(*directoryValue)(nil),
	(*directorySliceValue)(nil),
	(*existingFileValue)(nil),
	(*fileSliceValue)(nil),
	(*outputFileValue)(nil),
	(*urlValue)(nil),
	(*urlSliceValue)(nil),
}
