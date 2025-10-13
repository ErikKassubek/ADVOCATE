// Copyright (c) 2025 Erik Kassubek
//
// File: independentTraces.go
// Brief: Some traces are independent, meaning if one trace does not
//    contain a concurrency bug (panic or deadlock), the other cannot
//    contain one either. This file contains functions to check if two operations
//    are independent.
//
// Author: Erik Kassubek
// Created: 2025-10-13
//
// License: BSD-3-Clause

package guided

// TODO
