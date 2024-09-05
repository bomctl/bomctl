// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/utils/logger.go
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
package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

const levelWidth = 5

func New(prefix string) *log.Logger {
	// Set displayed width of log level in messages to show full level name
	styles := log.DefaultStyles()
	for _, level := range []log.Level{log.DebugLevel, log.ErrorLevel, log.FatalLevel, log.InfoLevel, log.WarnLevel} {
		styles.Levels[level] = styles.Levels[level].MaxWidth(levelWidth).Width(levelWidth)
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{Prefix: prefix, Level: log.GetLevel()})
	logger.SetStyles(styles)

	return logger
}
