// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/root.go
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
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const readWriteExecuteUser = 0o700

func initCache() {
	cacheDir := viper.GetString("cache_dir")

	if cache, err := os.UserCacheDir(); cacheDir == "" && err == nil {
		cacheDir = filepath.Join(cache, "bomctl")
		viper.Set("cache_dir", cacheDir)
	}

	cobra.CheckErr(os.MkdirAll(cacheDir, os.FileMode(readWriteExecuteUser)))
}

func initConfig() {
	cacheDir := viper.GetString("cache_dir")
	cfgFile := viper.GetString("config_file")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cfgDir, err := os.UserConfigDir()
		cobra.CheckErr(err)

		cfgDir = filepath.Join(cfgDir, "bomctl")
		cobra.CheckErr(os.MkdirAll(cfgDir, os.FileMode(readWriteExecuteUser)))

		viper.AddConfigPath(cfgDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("bomctl")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	viper.SetDefault("cache_dir", cacheDir)
}

func rootCmd() *cobra.Command {
	var verbose bool

	cobra.OnInitialize(initCache, initConfig)

	rootCmd := &cobra.Command{
		Use:     "bomctl",
		Long:    "Simpler Software Bill of Materials management",
		Version: Version,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	rootCmd.PersistentFlags().String("cache-dir", "",
		fmt.Sprintf("cache directory [defaults:\n\t%s\n\t%s\n\t%s",
			"Unix:    $HOME/.cache/bomctl",
			"Darwin:  $HOME/Library/Caches/bomctl",
			"Windows: %LocalAppData%\\bomctl]",
		),
	)

	rootCmd.PersistentFlags().String("config", "",
		fmt.Sprintf("config file [defaults:\n\t%s\n\t%s\n\t%s",
			"Unix:    $HOME/.config/bomctl/bomctl.yaml",
			"Darwin:  $HOME/Library/Application Support/bomctl/bomctl.yml",
			"Windows: %AppData%\\bomctl\\bomctl.yml]",
		),
	)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug output")

	// Bind flags to their associated viper configurations.
	cobra.CheckErr(viper.BindPFlag("cache_dir", rootCmd.PersistentFlags().Lookup("cache-dir")))
	cobra.CheckErr(viper.BindPFlag("config_file", rootCmd.PersistentFlags().Lookup("config")))
	cobra.CheckErr(viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")))

	rootCmd.AddCommand(fetchCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

func Execute() {
	cobra.CheckErr(rootCmd().Execute())
}
