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

	"github.com/bomctl/bomctl/internal/pkg/db"
)

var (
	cacheDir, cfgFile string
	logger            *log.Logger
	verbose           bool
)

func initCache() {
	if cache, err := os.UserCacheDir(); cacheDir == "" && err == nil {
		cacheDir = filepath.Join(cache, "bomctl")
	}

	cobra.CheckErr(os.MkdirAll(cacheDir, os.FileMode(0o700)))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cfgDir, err := os.UserConfigDir()
		cobra.CheckErr(err)

		cfgDir = filepath.Join(cfgDir, "bomctl")
		cobra.CheckErr(os.MkdirAll(cfgDir, os.FileMode(0o700)))

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
	cobra.OnInitialize(initCache, initConfig)

	rootCmd := &cobra.Command{
		Use:     "bomctl",
		Long:    "Simpler Software Bill of Materials management",
		Version: Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}

			_, err := db.CreateSchema(filepath.Join(cacheDir, "bomctl.db"))
			if err != nil {
				fmt.Fprintln(os.Stderr, "database creation: %w", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "",
		fmt.Sprintf("cache directory [defaults:\n\t%s\n\t%s\n\t%s",
			"Unix:    $HOME/.cache/bomctl",
			"Darwin:  $HOME/Library/Caches/bomctl",
			"Windows: %LocalAppData%\bomctl]",
		),
	)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file [defaults:\n\t%s\n\t%s\n\t%s",
			"Unix:    $HOME/.config/bomctl/bomctl.yaml",
			"Darwin:  $HOME/Library/Application Support/bomctl/bomctl.yml",
			"Windows: %AppData%\bomctl\bomctl.yml]",
		),
	)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug output")

	rootCmd.AddCommand(fetchCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

func Execute() {
	cobra.CheckErr(rootCmd().Execute())
}
