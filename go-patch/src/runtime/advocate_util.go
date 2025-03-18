// ADVOCATE-FILE-START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_util.go
// Brief: Helper functions
//
// Author: Erik Kassubek
// Created: 2023-05-25
//
// License: BSD-3-Clause

package runtime

import (
	"internal/runtime/atomic"
	"unsafe"
)

// MARK: INT -> STR

/*
 * Get a string representation of an uint64
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func uint64ToString(n uint64) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint64ToString(n/10) + string(rune(n%10+'0'))
	}
}

func pointerAddressAsString[T any](ptr *T, size bool) string {
	address := uintptr(unsafe.Pointer(ptr))

	// Handle zero case explicitly
	if address == 0 {
		return "0"
	}

	// Convert uintptr to string
	var str string
	for address > 0 {
		digit := address % 10         // Get the last digit
		str = string('0'+digit) + str // Prepend the digit
		address /= 10                 // Remove the last digit
	}

	if !size {
		return str
	}

	const desiredLength = 9

	// Get the length of the input string
	strLen := len(str)

	if strLen >= desiredLength {
		// If the string has 9 or more letters, return the last 9
		return str[strLen-desiredLength:]
	}

	return str
}

func pointerAddressAsUint32[T any](ptr *T, size bool) uint32 {
	return stringToUint32(pointerAddressAsString(ptr, size))
}

func pointerAddressAsUint64[T any](ptr *T, size bool) uint64 {
	return stringToUint64(pointerAddressAsString(ptr, size))
}

/*
 * Get a string representation of an int64
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func int64ToString(n int64) string {
	if n < 0 {
		return "-" + int64ToString(-n)
	}

	if n < 10 {
		return string(rune(n + '0'))
	}

	return int64ToString(n/10) + string(rune(n%10+'0'))
}

/*
 * Get a string representation of an int32
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func int32ToString(n int32) string {
	if n < 0 {
		return "-" + int32ToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return int32ToString(n/10) + string(rune(n%10+'0'))
	}
}

/*
 * Get a string representation of an uint32
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func uint32ToString(n uint32) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint32ToString(n/10) + string(rune(n%10+'0'))
	}
}

/*
 * Get a string representation of an int
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return intToString(n/10) + string(rune(n%10+'0'))

	}
}

// MARK: STR -> INT
/*
 * Convert a string to an integer
 * Works only with positive integers
 */
func stringToInt(s string) int {
	var result int
	sign := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			sign = -1
		} else if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + int(s[i]-'0')
		} else {
			panic("Invalid input")
		}
	}
	return result * sign
}

func stringToUint32(s string) uint32 {
	return uint32(stringToInt(s))
}

func stringToUint64(s string) uint64 {
	return uint64(stringToInt(s))
}

// MARK: BOOL -> STR

/*
 * Get a string representation of a bool
 * Args:
 * 	b: bool to convert
 * Return:
 * 	string representation of the bool (true: "t", false: "f")
 */
func boolToString(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

// MARK: STR manipulation

/*
 * Split a string at a separator
 * Args:
 * 	s: string to split
 * 	sep: separator
 * 	indices: at witch separators to split the string, must be sorted, 1 based
 * 		if nil split at all separators
 * Return:
 * 	split string
 */
func splitStringAtSeparator(s string, sep rune, indices []int) []string {
	var start int
	result := make([]string, 0)

	if indices == nil {
		for i, r := range s {
			if r == sep {
				result = append(result, s[start:i])
				start = i + 1
			}
		}
	} else {
		count := 0
		for _, index := range indices {
			for i, r := range s[start:] {
				if r == sep {
					count++
					if count == index {
						result = append(result, s[start:start+i])
						start += i + 1
						break
					}
				}
			}
		}
	}
	result = append(result, s[start:])
	return result
}

/*
 * Split a string at comma
 * Args:
 * 	s: string to split
 * 	indices: at witch commas to split the string, must be sorted, 1 based,
 * 		if nil split at all commas
 * Return:
 * 	splitted string
 */
func splitStringAtCommas(s string, indices []int) []string {
	return splitStringAtSeparator(s, ',', indices)
}

/*
 * Merge a string slice to a string separated by comma
 * Args:
 * 	s: slice of strings to merge
 * Return:
 * 	merged string, separated by commas
 */
func mergeString(s []string) string {
	return mergeStringSep(s, ",")
}

/*
 * Merge a string slice to a string
 * Args:
 * 	s: slice of strings to merge
 * 	sep: separator
 * Return:
 * 	merged string, separated by commas
 */
func mergeStringSep(s []string, sep string) string {
	var result string
	for i, elem := range s {
		if i != 0 {
			result += sep
		}
		result += elem
	}
	return result
}

/*
 * Split a string by the seperator
 */
func splitString(line string, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] == sep[0] {
			result = append(result, line[start:i])
			start = i + 1
		}
	}
	result = append(result, line[start:])
	return result
}

// MARK: ADVOCATE

var advocateCurrentRoutineID atomic.Uint64
var advocateGlobalCounter atomic.Uint64

/*
 * GetAdvocateRoutineID returns a new id for a routine
 * Return:
 * 	new id
 */
func GetAdvocateRoutineID() uint64 {
	id := advocateCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

/*
 * GetAdvocateObjectID returns a new id for a mutex, channel or waitgroup
 * Return:
 * 	new id
 */
func GetAdvocateObjectID() uint64 {
	routine := currentGoRoutine()

	if routine == nil {
		getg().advocateRoutineInfo = newAdvocateRoutine(getg())
		routine = currentGoRoutine()
	}

	routine.maxObjectId++
	if routine.maxObjectId > 999999999 {
		panic("Overflow Error: Tow many objects in one routine. Max: 999999999")
	}
	id := routine.id*1000000000 + routine.maxObjectId
	return id
}

/*
 * GetAdvocateCounter will update the timer and return the new value
 * Return:
 * 	new time value
 */
func GetNextTimeStep() uint64 {
	return advocateGlobalCounter.Add(2)
}

/*
 * Check if a list of integers contains an element
 * Args:
 * 	list: list of integers
 * 	elem: element to check
 * Return:
 * 	true if the list contains the element, false otherwise
 */
func containsInt(list []int, elem int) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	// Get the lengths of both the main string and the substring
	lenS := len(s)
	lenSub := len(sub)

	// If the substring is longer than the string, it can't be a substring
	if lenSub > lenS {
		return false
	}

	// Iterate over the main string `s`
	for i := 0; i <= lenS-lenSub; i++ {
		// Check if substring matches
		match := true
		for j := 0; j < lenSub; j++ {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		// If we found a match, return true
		if match {
			return true
		}
	}

	// No match found, return false
	return false
}


func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// ADVOCATE-FILE-END
