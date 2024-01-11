// -------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/root.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -------------------------------------------------------
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".bomctl")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func rootCmd() *cobra.Command {
	cobra.OnInitialize(initConfig)

	rootCmd := &cobra.Command{
		Use:     "bomctl",
		Long:    "Simpler Software Bill of Materials management",
		Version: getVersion(),
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bomctl.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug output")

	rootCmd.AddCommand(fetchCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

type (
	DirectoryValue      string
	ExistingFileValue   string
	OutputFileValue     string
	URLValue            string
	DirectorySliceValue []string
	FileSliceValue      []string
	URLSliceValue       []string
)

var (
	outputFile OutputFileValue
	sbomUrls   URLSliceValue
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
func (uv *URLValue) String() string             { return fmt.Sprintf("%v", *uv) }
func (usv *URLSliceValue) String() string       { return fmt.Sprintf("%v", *usv) }

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

func (uv *URLValue) Set(value string) error {
	*uv = URLValue(value)
	return nil
}

func (usv *URLSliceValue) Set(value string) error {
	*usv = append(*usv, value)
	return nil
}

const (
	valueTypeDir  string = "DIRECTORY"
	valueTypeFile string = "FILE"
	valueTypeURL  string = "URL"
)

func (dv *DirectoryValue) Type() string       { return valueTypeDir }
func (dsv *DirectorySliceValue) Type() string { return valueTypeDir }
func (efv *ExistingFileValue) Type() string   { return valueTypeFile }
func (fsv *FileSliceValue) Type() string      { return valueTypeFile }
func (ofv *OutputFileValue) Type() string     { return valueTypeFile }
func (uv *URLValue) Type() string             { return valueTypeURL }
func (usv *URLSliceValue) Type() string       { return valueTypeURL }

func Execute() {
	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
