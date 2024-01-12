// -------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/root.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -------------------------------------------------------
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cacheDir, cfgFile string

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
}

func rootCmd() *cobra.Command {
	cobra.OnInitialize(initCache, initConfig)

	rootCmd := &cobra.Command{
		Use:     "bomctl",
		Long:    "Simpler Software Bill of Materials management",
		Version: getVersion(),
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

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug output")

	rootCmd.AddCommand(fetchCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

func Execute() {
	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
