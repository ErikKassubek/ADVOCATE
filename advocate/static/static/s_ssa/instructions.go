// Copyright (c) 2026 Erik Kassubek
//
// File: instructions.go
// Brief: Instructions
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/ssa"
)

func (this *Data) analysisInstruction(f *Function, inst ssa.Instruction, i int) Instruction {

	var instru Instruction

	switch inst := inst.(type) {
	case *ssa.Alloc:
		instru = NewAlloc(f, inst, i)
	case *ssa.BinOp:
		instru = NewBinOp(f, inst, i)
	case *ssa.Call:
		instru = NewCall(f, inst, i)
	case *ssa.ChangeInterface:
		instru = NewChangeInterface(f, inst, i)
	case *ssa.ChangeType:
		instru = NewChangeType(f, inst, i)
	case *ssa.Convert:
		instru = NewConvert(f, inst, i)
	case *ssa.DebugRef:
		instru = NewDebugRef(f, inst, i)
	case *ssa.Defer:
		instru = NewDefer(f, inst, i)
	case *ssa.Extract:
		instru = NewExtract(f, inst, i)
	case *ssa.Field:
		instru = NewField(f, inst, i)
	case *ssa.FieldAddr:
		instru = NewFieldAddr(f, inst, i)
	case *ssa.Go:
		instru = NewGo(f, inst, i)
	case *ssa.If:
		instru = newIf(f, inst, i)
	case *ssa.Index:
		instru = NewIndex(f, inst, i)
	case *ssa.IndexAddr:
		instru = NewIndexAddr(f, inst, i)
	case *ssa.Jump:
		instru = newJump(f, inst, i)
	case *ssa.Lookup:
		instru = newLookup(f, inst, i)
	case *ssa.MakeChan:
		instru = newMakeChan(f, inst, i)
	case *ssa.MakeClosure:
		instru = newMakeClosure(f, inst, i)
	case *ssa.MakeInterface:
		instru = newMakeInterface(f, inst, i)
	case *ssa.MakeMap:
		instru = newMakeMap(f, inst, i)
	case *ssa.MakeSlice:
		instru = newMakeSlice(f, inst, i)
	case *ssa.MapUpdate:
		instru = newMapUpdate(f, inst, i)
	case *ssa.MultiConvert:
		instru = newMultiConvert(f, inst, i)
	case *ssa.Next:
		instru = newNext(f, inst, i)
	case *ssa.Panic:
		instru = newPanic(f, inst, i)
	case *ssa.Phi:
		instru = newPhi(f, inst, i)
	case *ssa.Range:
		instru = newRange(f, inst, i)
	case *ssa.Return:
		instru = newReturn(f, inst, i)
	case *ssa.RunDefers:
		instru = newRunDefers(f, inst, i)
	case *ssa.Select:
		instru = newSelect(f, inst, i)
	case *ssa.Send:
		instru = newSend(f, inst, i)
	case *ssa.Slice:
		instru = newSlice(f, inst, i)
	case *ssa.SliceToArrayPointer:
		instru = newSliceToArrayPointer(f, inst, i)
	case *ssa.Store:
		instru = newStore(f, inst, i)
	case *ssa.TypeAssert:
		instru = newTypeAssert(f, inst, i)
	case *ssa.UnOp:
		instru = newUnOp(f, inst, i)
	default:
		panic(fmt.Sprintf("Found unknown instruction: %s", inst.String()))
	}

	setContainsPrimitive(instru)

	instru.setRelevant(this)

	return instru
}

func setContainsPrimitive(inst Instruction) {
	hasChan := false
	hasMutex := false
	hasCond := false
	hasWaitGroup := false

	checkType := func(t types.Type) {
		if t == nil {
			return
		}

		if _, ok := t.Underlying().(*types.Chan); ok {
			hasChan = true
		}

		for {
			p, ok := t.(*types.Pointer)
			if !ok {
				break
			}
			t = p.Elem()
		}

		named, ok := t.(*types.Named)
		if !ok {
			return
		}

		obj := named.Obj()
		if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "sync" {
			return
		}

		switch obj.Name() {
		case "Mutex", "RWMutex":
			hasMutex = true
		case "Cond":
			hasCond = true
		case "WaitGroup":
			hasWaitGroup = true
		}
	}

	instr := inst.Inst()

	// Explicit channel instructions.
	switch i := instr.(type) {
	case *ssa.Alloc:
		hasChan = strings.HasPrefix(inst.Term(), "new chan")
	case *ssa.MakeChan, *ssa.Send, *ssa.Select:
		hasChan = true
	case *ssa.UnOp:
		if i.Op == token.ARROW {
			hasChan = true
		}
	}

	// Result value.
	if v, ok := instr.(ssa.Value); ok {
		checkType(v.Type())
	}

	// Operand values.
	for _, op := range instr.Operands(nil) {
		if op == nil || *op == nil {
			continue
		}
		checkType((*op).Type())
	}

	inst.setConc(hasConcInfo{hasChan, hasMutex, hasCond, hasWaitGroup})
}

func (this *Data) getInstructionPos(instr ssa.Instruction) (string, int) {
	pos := instr.Pos()
	if pos == token.NoPos {
		return "", 0
	}

	position := this.ssa.Fset.Position(pos)
	return position.Filename, position.Line
}
