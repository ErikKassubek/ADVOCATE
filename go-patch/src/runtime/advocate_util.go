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

// String
func buildTraceElemString(values ...any) string {
	res := ""
	for i, v := range values {
		if i != 0 {
			res += ","
		}

		res += convToString(v)
	}
	return res
}

func buildTraceElemStringSep(sep string, values ...any) string {
	res := ""
	for i, v := range values {
		if i != 0 {
			res += sep
		}

		res += convToString(v)
	}
	return res
}

func convToString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return intToString(v)
	case uint:
		return uint64ToString(uint64(v))
	case int32:
		return int32ToString(v)
	case int64:
		return int64ToString(v)
	case uint32:
		return uint32ToString(v)
	case uint64:
		return uint64ToString(v)
	case bool:
		if v {
			return "t"
		}
		return "f"
	}
	panic("unknown type")
	return ""
}

func posToString(file string, line int) string {
	return file + ":" + intToString(line)
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
