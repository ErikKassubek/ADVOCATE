// Copyrigth (c) 2024 Erik Kassubek
//
// File: utils.go
// Brief: Utility function to check if an slice contains a value
//
// Author: Erik Kassubek
// Created: 2024-04-06
//
// License: BSD-3-Clause

package utils

import (
	"fmt"
	"reflect"
	"strings"
)

/*
* Check if a slice ContainsString an element
* Args:
*   s: slice to check
*   e: element to check
* Returns:
*   bool: true is e in s, false otherwise
 */
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsInt(slice []int, elem int) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}

/*
 * Split the string into two parts at the last occurrence of the separator
 * Args:
 *   str (string): string to split
 *   sep (string): separator to split at
 * Returns:
 *   []string: If sep in string: list with two elements split at the sep,
 *     if not then list containing str
 */
func SplitAtLast(str string, sep string) []string {
	if sep == "" {
		return []string{str}
	}

	i := strings.LastIndex(str, sep)
	if i == -1 {
		return []string{str}
	}
	return []string{str[:i], str[i+1:]}
}

func GetErrorDiff(expected error, given error) error {
	if expected == nil && given == nil {
		return nil
	} else if expected == nil && given != nil {
		return fmt.Errorf("Given error is not nil. Expected nil")
	} else if expected != nil && given == nil {
		return fmt.Errorf("Given error is nil. Expected %s.", expected.Error())
	}

	// both are not nil
	if reflect.TypeOf(expected) != reflect.TypeOf(given) {
		return fmt.Errorf("Types of the errors are different.%T != %T", reflect.TypeOf(expected), reflect.TypeOf(given))
	}

	if expected.Error() != given.Error() {
		return fmt.Errorf("Errors contain different messages. %s != %s.", expected.Error(), given.Error())
	}

	return nil
}

func AddIfNotContains(l []int, e int) []int {
	if !ContainsInt(l, e) {
		l = append(l, e)
	}
	return l
}

/*
 * Given two lists of ints, return a list containing all the elements from both
 * lists. The resulting list does not contain duplicated.
 */
func MergeIntLists(l1, l2 []int) []int {
	uniqueMap := make(map[int]bool)
	res := []int{}

	for _, val := range l1 {
		if !uniqueMap[val] {
			uniqueMap[val] = true
			res = append(res, val)
		}
	}

	for _, val := range l2 {
		if !uniqueMap[val] {
			uniqueMap[val] = true
			res = append(res, val)
		}
	}

	return res
}
