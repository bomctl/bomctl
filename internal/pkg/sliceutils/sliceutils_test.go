// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/sliceutils/sliceutils_test.go
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

package sliceutils_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/sliceutils"
)

type sliceutilsSuite struct {
	suite.Suite
}

func (ss *sliceutilsSuite) TestAll() {
	for _, data := range []struct {
		cond     func(any) bool
		items    []any
		expected bool
	}{
		{
			cond:     nil,
			items:    []any{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			cond:     nil,
			items:    []any{6, 7, 8, 9, 0},
			expected: false,
		},
		{
			cond: func(item any) bool {
				intItem, ok := item.(int)

				return ok && intItem > 0
			},
			items:    []any{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			cond: func(item any) bool {
				intItem, ok := item.(int)

				return ok && intItem > 0
			},
			items:    []any{6, 7, 8, 9, 0},
			expected: false,
		},
	} {
		actual := sliceutils.All(data.items, data.cond)

		ss.Require().Equal(data.expected, actual)
	}
}

func (ss *sliceutilsSuite) TestAny() {
	for _, data := range []struct {
		cond     func(any) bool
		items    []any
		expected bool
	}{
		{
			cond:     nil,
			items:    []any{"", "one", "two", "three"},
			expected: true,
		},
		{
			cond:     nil,
			items:    []any{"", "", "", ""},
			expected: false,
		},
		{
			cond: func(item any) bool {
				strItem, ok := item.(string)

				return ok && strings.HasPrefix(strItem, "t")
			},
			items:    []any{"", "one", "two", "three"},
			expected: true,
		},
		{
			cond: func(item any) bool {
				strItem, ok := item.(string)

				return ok && strings.HasPrefix(strItem, "s")
			},
			items:    []any{"one", "two", "", "four"},
			expected: false,
		},
	} {
		actual := sliceutils.Any(data.items, data.cond)

		ss.Require().Equal(data.expected, actual)
	}
}

func (ss *sliceutilsSuite) TestExtract() {
	type testStringValue struct {
		value string
	}

	for _, data := range []struct {
		cond     func(testStringValue) string
		items    []testStringValue
		expected []string
	}{
		{
			cond:     func(item testStringValue) string { return item.value },
			items:    []testStringValue{{""}, {"one"}, {"two"}, {"three"}},
			expected: []string{"", "one", "two", "three"},
		},
	} {
		actual := sliceutils.Extract(data.items, data.cond)

		ss.Require().Equal(data.expected, actual)
	}
}

func (ss *sliceutilsSuite) TestFilter() {
	type testStringValue struct {
		value string
	}

	for _, data := range []struct {
		cond     func(testStringValue) bool
		items    []testStringValue
		expected []testStringValue
	}{
		{
			cond:     nil,
			items:    []testStringValue{{""}, {"one"}, {"two"}, {"three"}},
			expected: []testStringValue{{"one"}, {"two"}, {"three"}},
		},
		{
			cond:     nil,
			items:    []testStringValue{{""}, {""}, {""}, {""}},
			expected: []testStringValue{},
		},
		{
			cond: func(item testStringValue) bool {
				return strings.HasPrefix(item.value, "t")
			},
			items:    []testStringValue{{""}, {"one"}, {"two"}, {"three"}},
			expected: []testStringValue{{"two"}, {"three"}},
		},
	} {
		actual := sliceutils.Filter(data.items, data.cond)

		ss.Require().Equal(data.expected, actual)
	}
}

func (ss *sliceutilsSuite) TestMap() {
	{
		type testIntValue struct {
			value int
		}

		items := []testIntValue{{0}, {1}, {2}, {3}, {4}, {5}}
		expected := []testIntValue{{0}, {2}, {4}, {6}, {8}, {10}}

		actual := sliceutils.Map(items, func(item testIntValue) testIntValue {
			item.value *= 2

			return item
		})

		ss.Require().Equal(expected, actual)
	}

	{
		type testStringValue struct {
			value string
		}

		items := []testStringValue{{"one"}, {"two"}, {"three"}, {"four"}}
		expected := []testStringValue{{"one-mapped"}, {"two-mapped"}, {"three-mapped"}, {"four-mapped"}}

		actual := sliceutils.Map(items, func(item testStringValue) testStringValue {
			item.value += "-mapped"

			return item
		})

		ss.Require().Equal(expected, actual)
	}
}

func (ss *sliceutilsSuite) TestNext() {
	type testAnyValue struct {
		value any
	}

	for _, data := range []struct {
		cond      func(testAnyValue) bool
		expected  testAnyValue
		items     []testAnyValue
		shouldErr bool
	}{
		{
			cond:     nil,
			items:    []testAnyValue{{nil}, {"one"}, {"two"}, {"three"}},
			expected: testAnyValue{"one"},
		},
		{
			cond:      nil,
			items:     []testAnyValue{{nil}, {nil}, {nil}, {nil}},
			shouldErr: true,
		},
		{
			cond:      nil,
			items:     []testAnyValue{{[]any{nil}}, {[]any{nil}}, {[]any{nil}}, {[]any{nil}}},
			shouldErr: true,
		},
		{
			cond: func(item testAnyValue) bool {
				return item.value != ""
			},
			items:     []testAnyValue{{""}, {""}, {""}, {""}},
			shouldErr: true,
		},
		{
			cond: func(item testAnyValue) bool {
				return item.value != ""
			},
			items:    []testAnyValue{{""}, {"two"}, {""}, {""}},
			expected: testAnyValue{"two"},
		},
	} {
		actual, err := sliceutils.Next(data.items, data.cond)

		if data.shouldErr {
			ss.Error(err)
		} else {
			ss.Equal(data.expected, actual)
		}
	}
}

func TestSliceUtilsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(sliceutilsSuite))
}
