// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/logger"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

const (
	modeUserRead  = 0o400
	modeUserWrite = 0o200
	modeUserExec  = 0o100
	cliTableWidth = 80
)

type optionsKey struct{}

func backendFromContext(cmd *cobra.Command) *db.Backend {
	backend, err := db.BackendFromContext(cmd.Context())
	if err != nil {
		logger.New("").Fatal(err)
	}

	return backend
}

func defaultCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	cobra.CheckErr(err)

	return filepath.Join(cacheDir, "bomctl")
}

func defaultConfig() string {
	cfgDir, err := os.UserConfigDir()
	cobra.CheckErr(err)

	return filepath.Join(cfgDir, "bomctl", "bomctl.yaml")
}

func initCache() {
	cacheDir := viper.GetString("cache_dir")
	cobra.CheckErr(os.MkdirAll(cacheDir, modeUserRead|modeUserWrite|modeUserExec))
}

func initConfig(cmd *cobra.Command) func() {
	return func() {
		cfgFile := cmd.PersistentFlags().Lookup("config").Value.String()

		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			cfgDir, err := os.UserConfigDir()
			cobra.CheckErr(err)

			cfgDir = filepath.Join(cfgDir, "bomctl")
			cobra.CheckErr(os.MkdirAll(cfgDir, modeUserRead|modeUserWrite|modeUserExec))

			viper.AddConfigPath(cfgDir)
			viper.SetConfigType("yaml")
			viper.SetConfigName("bomctl")
		}

		viper.SetEnvPrefix("bomctl")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err == nil {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}

		cobra.CheckErr(os.MkdirAll(viper.GetString("cache_dir"), modeUserRead|modeUserWrite|modeUserExec))
	}
}

func optionsFromContext(cmd *cobra.Command) *options.Options {
	opts, ok := cmd.Context().Value(optionsKey{}).(*options.Options)
	if !ok {
		logger.New("").Fatal("Failed to get options from command context")
	}

	return opts
}

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bomctl",
		Long:    "Simpler Software Bill of Materials management",
		Version: VersionString,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			if verbosity > 0 {
				log.SetLevel(log.DebugLevel)
			}

			cacheDir, err := cmd.Flags().GetString("cache-dir")
			cobra.CheckErr(err)

			// Get first top-level subcommand.
			subcmd := cmd
			for subcmd.HasParent() && subcmd.Parent() != subcmd.Root() {
				subcmd = subcmd.Parent()
			}

			opts := options.New().
				WithCacheDir(cacheDir).
				WithConfigFile(viper.ConfigFileUsed()).
				WithVerbosity(verbosity).
				WithLogger(logger.New(subcmd.Name()))

			backend, err := db.NewBackend(
				db.WithDatabaseFile(filepath.Join(cacheDir, db.DatabaseFile)),
				db.WithVerbosity(verbosity),
			)
			if err != nil {
				opts.Logger.Fatalf("%v", err)
			}

			cmd.SetContext(context.WithValue(cmd.Context(), optionsKey{}, opts))
			cmd.SetContext(context.WithValue(cmd.Context(), db.BackendKey{}, backend))
			opts.WithContext(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().String("cache-dir", defaultCacheDir(), "cache directory")
	rootCmd.PersistentFlags().String("config", defaultConfig(), "config file")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Enable debug output")

	cobra.OnInitialize(initCache, initConfig(rootCmd))

	// Bind flags to their associated viper configurations.
	cobra.CheckErr(viper.BindPFlag("cache_dir", rootCmd.PersistentFlags().Lookup("cache-dir")))

	rootCmd.AddCommand(exportCmd())
	rootCmd.AddCommand(fetchCmd())
	rootCmd.AddCommand(importCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(mergeCmd())
	rootCmd.AddCommand(pushCmd())
	rootCmd.AddCommand(aliasCmd())
	rootCmd.AddCommand(tagCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

func Execute() {
	cobra.CheckErr(rootCmd().Execute())
}
