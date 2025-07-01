// Copyright (c) 2025 Erik Kassubek
//
// File: constants.go
// Brief: Constants that can be set via the setting flag
//
// Author: Erik Kassubek
// Created: 2025-05-06
//
// License: BSD-3-Clause

package helper

import (
	"advocate/utils/log"
	"math"
	"strconv"
	"strings"
)

var (
	GFuzzW1 = 10.0
	GFuzzW2 = 10.0
	GFuzzW3 = 10.0
	GFuzzW4 = 10.0

	GFuzzFlipP    = 0.99
	GFuzzFlipPMin = 0.1

	GoPieW1        = 1.0
	GoPieW2        = 1.0
	GoPieBound     = 3
	GoPieMutabound = 128

	GoPieSCStart = 5
)

// SetSettings sets different constants and settings used in the program
// from the settings flag
//
// Parameter
//   - settings string: settings string of the form `name1=value1,name2=value2,...`
//   - maxFuzzingRun int: maxFuzzingRun flag value
//   - fuzzingMode string: fuzzingMode flag value
func SetSettings(settings string, maxFuzzingRun int, fuzzingMode string) {
	if fuzzingMode != "GoPie" {
		GoPieMutabound = min(int(maxFuzzingRun/GoPieSCStart), 128)
	}

	if settings == "" {
		return
	}

	sets := strings.Split(settings, ",")

	for _, elem := range sets {
		elemParts := strings.Split(elem, "=")
		if len(elemParts) != 2 {
			log.Errorf("Invalid setting '%s'. Skip.", elem)
			continue
		}

		value, err := strconv.ParseFloat(elemParts[1], 64)
		if err != nil {
			log.Errorf("Value in %s cannot be converted to number. Skip.", elem)
			continue
		}

		switch elemParts[0] {
		case "GFuzzW1":
			GFuzzW1 = value
		case "GFuzzW2":
			GFuzzW2 = value
		case "GFuzzW3":
			GFuzzW3 = value
		case "GFuzzW4":
			GFuzzW4 = value
		case "GFuzzFlipP":
			GFuzzFlipP = clamp(value, 0, 1)
		case "GFuzzFlipPMin":
			GFuzzFlipPMin = clamp(value, 0, 1)
		case "GoPieW1":
			GoPieW1 = value
		case "GoPieW2":
			GoPieW2 = value
		case "GoPieBound":
			GoPieBound = int(clamp(value, 2.0, math.MaxFloat64))
		case "GoPieMutabound":
			GoPieMutabound = int(clamp(value, 1, math.MaxFloat64))
		case "GoPieSCStart":
			GoPieSCStart = int(clamp(value, 1, math.MaxFloat64))
		default:
			log.Errorf("Unknown name in setting %s", elemParts[0])
			continue
		}
	}
}

// Return the value. If it is bigger than maxVal, return maxVal. If it is smaller
// than minVal, return minVal
//
// Parameter:
//   - value float64: the value
//   - minVal flat64: the minimum value
//   - maxVal flat64: the maximum value
//
// Returns:\
//   - the value clamped between minVal and maxVal
func clamp(value, minVal, maxVal float64) float64 {
	if value >= maxVal {
		return maxVal
	} else if value <= minVal {
		return minVal
	}
	return value
}
