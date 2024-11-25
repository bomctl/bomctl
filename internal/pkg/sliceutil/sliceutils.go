// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/sliceutil/sliceutils.go
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

package sliceutil

import (
	"errors"
	"reflect"
)

var ErrItemsExhausted = errors.New("no match found in items")

// All returns true if all items satisfy the specified condition.
// If cond is nil, checks whether all items are a value other than zero value (truthy).
//
// # Example:
//
//	func main() {
//		items := []int{0, 1, 2, 3}
//
//		fmt.Println(sliceutils.All(items, nil))
//		fmt.Println(sliceutils.All(items, func(item int) bool { return item < 5 }))
//
//		// Output:
//		// false
//		// true
//	}
func All[T any](items []T, cond func(T) bool) bool {
	if cond == nil {
		cond = isTruthy
	}

	for idx := range items {
		if !cond(items[idx]) {
			return false
		}
	}

	return true
}

// Any returns true if any item satisfies the specified condition.
// If cond is nil, checks whether any item is a value other than zero value (truthy).
//
// # Example:
//
//	func main() {
//		items := []int{0, 0, -1, 0}
//
//		fmt.Println(sliceutils.Any(items, nil))
//		fmt.Println(sliceutils.Any(items, func(item int) bool { return item > 0 }))
//
//		// Output:
//		// true
//		// false
//	}
func Any[T any](items []T, cond func(T) bool) bool {
	if cond == nil {
		cond = isTruthy
	}

	for idx := range items {
		if cond(items[idx]) {
			return true
		}
	}

	return false
}

// Extract extracts fields from a slice of structs.
//
// # Example:
//
//	type toExtract struct {
//		value int
//	}
//
//	func main() {
//		filter := func(et toExtract) int {
//			return et.value
//		}
//
//		fmt.Printf("%+v\n", sliceutils.Extract([]toExtract{{1}, {2}, {3}, {4}}, filter))
//
//		// Output:
//		// [1 2 3 4]
//	}
func Extract[T any, E any](items []T, cond func(T) E) []E {
	extracted := make([]E, 0, len(items))

	for idx := range items {
		extracted = append(extracted, cond(items[idx]))
	}

	return extracted
}

// Filter returns a slice of items from the original slice that satisfy the specified condition.
// If cond is nil, items equal to the underlying type's zero value (falsy) are filtered out.
//
// # Example:
//
//	type toFilter struct {
//		value int
//	}
//
//	func main() {
//		filter := func(tf toFilter) bool {
//			return tf.value > 0
//		}
//
//		fmt.Printf("%+v\n", sliceutils.Filter([]toFilter{{0}, {1}, {2}, {3}}, filter))
//		fmt.Printf("%+v\n", sliceutils.Filter([]string{"one", "two", "", "three"}, nil))
//
//		// Output:
//		// [{value:1} {value:2} {value:3}]
//		// [one two three]
//	}
func Filter[T any](items []T, cond func(T) bool) []T {
	filtered := make([]T, 0, len(items))

	if cond == nil {
		cond = isTruthy
	}

	for idx := range items {
		if cond(items[idx]) {
			filtered = append(filtered, items[idx])
		}
	}

	return filtered
}

// Map applies a transformation function to each item of a slice.
//
// # Example:
//
//	type toMap struct {
//		value int
//	}
//
//	func main() {
//		filter := func(et toMap) toMap {
//			et.value *= 2
//
//			return et
//		}
//
//		fmt.Printf("%+v\n", sliceutils.Map([]toMap{{0}, {1}, {2}, {3}, {4}}, filter))
//
//		// Output:
//		// [{value:0} {value:2} {value:4} {value:6} {value:8}]
//	}
func Map[T any](items []T, mapFunc func(T) T) []T {
	updated := make([]T, 0, len(items))

	for idx := range items {
		updated = append(updated, mapFunc(items[idx]))
	}

	return updated
}

// Next returns the next item in a slice that satisifies a given condition.
// If cond is nil, returns the next item with a value other than zero value (truthy).
//
// # Example:
//
//	func main() {
//		filter := func(item struct{ int }) bool {
//			return item.int > 2
//		}
//
//		if nextItem, err := sliceutils.Next([]struct{ int }{{0}, {1}, {2}, {3}, {4}}, filter); err == nil {
//			fmt.Printf("%+v\n", nextItem)
//		}
//
//		if _, err := sliceutils.Next([]struct{ int }{{0}, {-1}, {-2}, {-3}, {-4}}, filter); err != nil {
//			fmt.Printf("%+v\n", err)
//		}
//
//		if _, err := sliceutils.Next([]struct{ int }{{0}, {0}, {0}, {0}, {0}}, nil); err != nil {
//			fmt.Printf("%+v\n", err)
//		}
//
//		// Output:
//		// {int:3}
//		// no match found in items
//		// no match found in items
//	}
func Next[T any](items []T, cond func(T) bool) (T, error) {
	if cond == nil {
		cond = isTruthy
	}

	for idx := range items {
		if cond(items[idx]) {
			return items[idx], nil
		}
	}

	var unset T

	return unset, ErrItemsExhausted
}

func isTruthy[T any](item T) bool {
	value := reflect.ValueOf(item)
	if value.Comparable() {
		zeroValue := reflect.Zero(reflect.TypeOf(item))

		return !value.Equal(zeroValue)
	}

	return false
}
