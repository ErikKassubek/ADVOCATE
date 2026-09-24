// Copyright (c) 2026 Erik Kassubek
//
// File: block.go
// Brief: Block
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"fmt"

	"golang.org/x/tools/go/ssa"
)

type Block struct {
	id    int
	insts []Instruction
}

func (self *Block) String() string {
	res := fmt.Sprintf("%d:\n", self.id)
	for _, inst := range self.insts {
		res += "\t\t" + inst.StringInfo() + "\n"
	}
	return res
}

func (this *Block) Instrs() []Instruction {
	if this == nil {
		return make([]Instruction, 0)
	}
	return this.insts
}

func (this *Data) analysisBlock(f *Function, bl *ssa.BasicBlock) *Block {
	b := Block{
		id:    bl.Index,
		insts: make([]Instruction, len(bl.Instrs)),
	}

	for i, inst := range bl.Instrs {
		b.insts[i] = this.analysisInstruction(f, inst, i)
	}

	return &b
}
