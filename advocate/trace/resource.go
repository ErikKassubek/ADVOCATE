// Copyright (c) 2024 Erik Kassubek
//
// File: resources.go
// Brief: Resources
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

// ========================================================
// MARK: Resource
// ========================================================

type Resource struct {
	id    int
	alloc *ElementAlloc
}

func NewResource(id int, alloc *ElementAlloc) Resource {
	return Resource{id: id, alloc: alloc}
}

func (this *Resource) Id() int {
	return this.id
}

func (this *Resource) Alloc() *ElementAlloc {
	return this.alloc
}
