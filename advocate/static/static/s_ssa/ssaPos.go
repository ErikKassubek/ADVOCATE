// Copyright (c) 2026 Erik Kassubek
//
// File: ssaPos.go
// Brief: Store a position in the ssa
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

type SsaPos struct {
	F      *Function
	B      *Block
	I      Instruction
	InstID int
}

func NewSsaPos(f *Function, b *Block, i Instruction, instID int) SsaPos {
	return SsaPos{F: f, B: b, I: i, InstID: instID}
}

func NewSsaPosFunc(f *Function) SsaPos {
	b := f.Blocks()[0]
	i := b.Instrs()[0]
	return SsaPos{F: f, B: b, I: i, InstID: 0}
}

func NewSsaPosFuncBlock(f *Function, blockID int) SsaPos {
	b := f.Blocks()[blockID]
	i := b.Instrs()[0]
	return SsaPos{F: f, B: b, I: i, InstID: 0}
}

func NewNilSsaPos() SsaPos {
	return SsaPos{nil, nil, nil, 0}
}

func (this *SsaPos) Blocks() []*Block {
	return this.F.Blocks()
}

func (this *SsaPos) Instrs() []Instruction {
	return this.B.Instrs()
}

func (this *SsaPos) Nil() bool {
	return this.F == nil
}

func (this *SsaPos) NewBlock(b *Block) {
	this.B = b
	this.InstID = 0
	this.I = this.B.Instrs()[0]
}

func (this SsaPos) String() string {
	f := this.F.Name()
	if f == "" {
		f = "init"
	}
	return f + " : " + this.I.String()
}

func (this *SsaPos) Next() SsaPos {
	instID := this.InstID + 1
	i := this.B.Instrs()[instID]
	res := SsaPos{F: this.F, B: this.B, I: i, InstID: instID}
	return res
}
