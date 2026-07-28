// Copyright (c) 2026 Erik Kassubek
//
// File: instructions.go
// Brief: Instructions
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/static/static/s_base"
	"advocate/utils/log"
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

type InstClass string

const (
	Ic_unknown             InstClass = "unknown"
	Ic_alloc               InstClass = "alloc"
	Ic_binOp               InstClass = "binOp"
	Ic_call                InstClass = "call"
	Ic_changeInterface     InstClass = "changeInterface"
	Ic_changeType          InstClass = "changeType"
	Ic_const               InstClass = "const"
	Ic_convert             InstClass = "convert"
	Ic_debugRef            InstClass = "debugRef"
	Ic_defer               InstClass = "defer"
	Ic_extract             InstClass = "extract"
	Ic_field               InstClass = "field"
	Ic_fieldAddr           InstClass = "fieldAddr"
	Ic_freeVar             InstClass = "freeVar"
	Ic_function            InstClass = "function"
	Ic_go                  InstClass = "go"
	Ic_if                  InstClass = "if"
	Ic_index               InstClass = "index"
	Ic_indexAddr           InstClass = "indexAddr"
	Ic_jump                InstClass = "jump"
	Ic_lookup              InstClass = "lookup"
	Ic_makeChan            InstClass = "makeChan"
	Ic_makeClosure         InstClass = "makeClosure"
	Ic_makeInterface       InstClass = "makeInterface"
	Ic_makeMap             InstClass = "makeMap"
	Ic_makeSlice           InstClass = "makeSlice"
	Ic_mapUpdate           InstClass = "mapUpdate"
	Ic_multiConvert        InstClass = "multiConvert"
	Ic_next                InstClass = "next"
	Ic_panic               InstClass = "panic"
	Ic_parameter           InstClass = "parameter"
	Ic_phi                 InstClass = "phi"
	Ic_range               InstClass = "range"
	Ic_return              InstClass = "return"
	Ic_runDefers           InstClass = "runDefers"
	Ic_select              InstClass = "select"
	Ic_send                InstClass = "send"
	Ic_slice               InstClass = "slice"
	Ic_sliceToArrayPointer InstClass = "sliceToArrayPointer"
	Ic_store               InstClass = "store"
	Ic_typeAssert          InstClass = "typeAssert"
	Ic_unOp                InstClass = "unOp"
)

type Instruction interface {
	Name() string
	String() string
	StringInfo() string
	Inst() ssa.Instruction
	InTrace() bool
	Relevant() bool

	Conc() hasConcInfo
	HasChannel() bool
	HasMutex() bool
	HasCond() bool
	HasWG() bool

	Class() InstClass

	setName(name string)
	setRelevant(rel, trace bool)
	setInst(inst ssa.Instruction)
	setConc(conc hasConcInfo)
}

// ================================================================
// MARK: Base
// ================================================================

type InstructionBase struct {
	name string
	inst ssa.Instruction

	inTrace bool
	relvant bool

	class InstClass

	conc hasConcInfo
}

func (this *InstructionBase) String() (res string) {
	if this.name == "" {
		res = this.inst.String()
	} else {
		res = fmt.Sprintf("%s = %s", this.name, this.inst.String())
	}

	return
}

func (this *InstructionBase) StringInfo() (res string) {
	if this.relvant {
		res += "+"
	} else {
		res += "-"
	}

	if this.inTrace {
		res += "+ "
	} else {
		res += "- "
	}

	res += fmt.Sprintf("%-40s", this.String())

	// name
	obj := func(i int) s_base.ObjName {
		switch concRes(i) {
		case chanInd:
			return s_base.Channel
		case mutexInd:
			return s_base.Mutex
		case condVarInd:
			return s_base.CondVar
		case wgInd:
			return s_base.Wg
		default:
			return s_base.UnknownObj
		}
	}

	if this.class != Ic_unknown {
		res += "\t-> " + fmt.Sprintf("%-20s", string(this.class))
	}

	found := false
	for i := 0; i < 4; i++ {
		if this.conc[i] {
			if !found {
				res += "  -> "
			} else {
				res += ", "
			}
			res += string(obj(i))
			found = true
		}
	}

	return
}

func (this *InstructionBase) Name() string {
	return this.name
}

func (this *InstructionBase) Class() InstClass {
	return this.class
}

func (this *InstructionBase) Inst() ssa.Instruction {
	return this.inst
}

func (this *InstructionBase) Conc() hasConcInfo {
	return this.conc
}

func (this *InstructionBase) HasChannel() bool {
	return this.conc[chanInd]
}

func (this *InstructionBase) HasMutex() bool {
	return this.conc[mutexInd]
}

func (this *InstructionBase) HasCond() bool {
	return this.conc[condVarInd]
}

func (this *InstructionBase) HasWG() bool {
	return this.conc[wgInd]
}

func (this *InstructionBase) Relevant() bool {
	return this.relvant
}

func (this *InstructionBase) InTrace() bool {
	return this.inTrace
}

func (this *InstructionBase) setRelevant(rel, trace bool) {
	this.relvant = rel
	this.inTrace = trace
}

func (this *InstructionBase) setName(name string) {
	this.name = name
}

func (this *InstructionBase) setInst(inst ssa.Instruction) {
	this.inst = inst
}

func (this *InstructionBase) setConc(conc hasConcInfo) {
	this.conc = conc
}

func (this *Data) analysisInstruction(instr ssa.Instruction) Instruction {

	var inst Instruction

	switch instr := instr.(type) {
	case *ssa.Alloc:
		inst = &InstructionAlloc{InstructionBase: InstructionBase{class: Ic_alloc}}
	case *ssa.BinOp:
		inst = &InstructionBinOp{InstructionBase: InstructionBase{class: Ic_binOp}}
	case *ssa.Call:
		inst = NewInstructionCall(instr)
	case *ssa.ChangeInterface:
		inst = &InstructionChangeInterface{InstructionBase: InstructionBase{class: Ic_changeInterface}}
	case *ssa.ChangeType:
		inst = &InstructionChangeType{InstructionBase: InstructionBase{class: Ic_changeType}}
	case *ssa.Convert:
		inst = &InstructionConvert{InstructionBase: InstructionBase{class: Ic_convert}}
	case *ssa.DebugRef:
		inst = &InstructionDebugRef{InstructionBase: InstructionBase{class: Ic_debugRef}}
	case *ssa.Defer:
		inst = &InstructionDefer{InstructionBase: InstructionBase{class: Ic_defer}}
	case *ssa.Extract:
		inst = &InstructionExtract{InstructionBase: InstructionBase{class: Ic_extract}}
	case *ssa.Field:
		inst = &InstructionField{InstructionBase: InstructionBase{class: Ic_field}}
	case *ssa.FieldAddr:
		inst = &InstructionFieldAddr{InstructionBase: InstructionBase{class: Ic_fieldAddr}}
	case *ssa.Go:
		inst = &InstructionGo{InstructionBase: InstructionBase{class: Ic_go}}
	case *ssa.If:
		inst = newInstructionIf(instr)
	case *ssa.Index:
		inst = &InstructionIndex{InstructionBase: InstructionBase{class: Ic_index}}
	case *ssa.IndexAddr:
		inst = &InstructionIndexAddr{InstructionBase: InstructionBase{class: Ic_indexAddr}}
	case *ssa.Jump:
		inst = newInstructionJump(instr)
	case *ssa.Lookup:
		inst = &InstructionLookup{InstructionBase: InstructionBase{class: Ic_lookup}}
	case *ssa.MakeChan:
		inst = &InstructionMakeChan{InstructionBase: InstructionBase{class: Ic_makeChan}}
	case *ssa.MakeClosure:
		inst = &InstructionMakeClosure{InstructionBase: InstructionBase{class: Ic_makeClosure}}
	case *ssa.MakeInterface:
		inst = &InstructionMakeInterface{InstructionBase: InstructionBase{class: Ic_makeInterface}}
	case *ssa.MakeMap:
		inst = &InstructionMakeMap{InstructionBase: InstructionBase{class: Ic_makeMap}}
	case *ssa.MakeSlice:
		inst = &InstructionMakeSlice{InstructionBase: InstructionBase{class: Ic_makeSlice}}
	case *ssa.MapUpdate:
		inst = &InstructionMapUpdate{InstructionBase: InstructionBase{class: Ic_mapUpdate}}
	case *ssa.MultiConvert:
		inst = &InstructionMultiConvert{InstructionBase: InstructionBase{class: Ic_multiConvert}}
	case *ssa.Next:
		inst = &InstructionNext{InstructionBase: InstructionBase{class: Ic_next}}
	case *ssa.Panic:
		inst = &InstructionPanic{InstructionBase: InstructionBase{class: Ic_panic}}
	case *ssa.Phi:
		inst = &InstructionPhi{InstructionBase: InstructionBase{class: Ic_phi}}
	case *ssa.Range:
		inst = &InstructionRange{InstructionBase: InstructionBase{class: Ic_range}}
	case *ssa.Return:
		inst = &InstructionReturn{InstructionBase: InstructionBase{class: Ic_return}}
	case *ssa.RunDefers:
		inst = &InstructionRunDefers{InstructionBase: InstructionBase{class: Ic_runDefers}}
	case *ssa.Select:
		inst = &InstructionSelect{InstructionBase: InstructionBase{class: Ic_select}}
	case *ssa.Send:
		inst = &InstructionSend{InstructionBase: InstructionBase{class: Ic_send}}
	case *ssa.Slice:
		inst = &InstructionSlice{InstructionBase: InstructionBase{class: Ic_slice}}
	case *ssa.SliceToArrayPointer:
		inst = &InstructionSliceToArrayPointer{InstructionBase: InstructionBase{class: Ic_sliceToArrayPointer}}
	case *ssa.Store:

		inst = &InstructionStore{InstructionBase: InstructionBase{class: Ic_store}}
	case *ssa.TypeAssert:
		inst = &InstructionTypeAssert{InstructionBase: InstructionBase{class: Ic_typeAssert}}
	case *ssa.UnOp:
		inst = &InstructionUnOp{InstructionBase: InstructionBase{class: Ic_unOp}}
	default:
		log.Error("Found unknown instruction: ", instr.String())
		inst = &InstructionUnknown{InstructionBase: InstructionBase{class: Ic_unknown}}
	}

	name := ""
	switch v := instr.(type) {
	case ssa.Value:
		name = v.Name()
	case nil:
		// Be robust against bad transforms.
		name = "<deleted>"
	}
	inst.setName(name)

	inst.setInst(instr)

	setContainsPrimitive(inst)

	this.isRelvant(inst)

	return inst
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

// Determine if instruction is relevant
//
// Parameter:
//   - instr ssa.Instruction: the instruciton
//
// Returns:rele
//   - bool: is relevant
//   - bool: is in trace
func (this *Data) isRelvant(instr Instruction) {
	resource := instr.Conc().Resource()

	switch instr := instr.(type) {
	case *InstructionGo, *InstructionMakeChan, *InstructionReturn, *InstructionSelect, *InstructionSend, *InstructionIf:
		instr.setRelevant(true, true)
	case *InstructionAlloc, *InstructionUnOp:
		instr.setRelevant(resource, resource)
	case *InstructionMapUpdate, *InstructionLookup, *InstructionExtract, *InstructionStore:
		instr.setRelevant(resource, false)
	case *InstructionField,
		*InstructionFieldAddr,
		*InstructionFreeVar,
		*InstructionIndex,
		*InstructionIndexAddr,
		*InstructionJump,
		*InstructionMakeClosure,
		*InstructionMakeInterface,
		*InstructionMakeMap,
		*InstructionMakeSlice,
		*InstructionRange,
		*InstructionNext,
		*InstructionRunDefers,
		*InstructionSlice,
		*InstructionSliceToArrayPointer:
		instr.setRelevant(true, false)
	case *InstructionCall:
		i := instr.Inst().(*ssa.Call)
		fn := i.Common().StaticCallee()

		if fn == nil {
			instr.setRelevant(resource, resource)
			return
		}

		if _, ok := this.funcs[fn.String()]; !ok {
			instr.setRelevant(resource, resource)
			return
		}

		instr.setRelevant(true, true)
	default:
		instr.setRelevant(false, false)
	}
}

func (this *Data) getInstructionPos(instr ssa.Instruction) (string, int) {
	pos := instr.Pos()
	if pos == token.NoPos {
		return "", 0
	}

	position := this.ssa.Fset.Position(pos)
	return position.Filename, position.Line
}

// ================================================================
// MARK: InstructionUnknown
// ================================================================

type InstructionUnknown struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionAlloc
// ================================================================

type InstructionAlloc struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionBinOp
// ================================================================

type InstructionBinOp struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionCall
// ================================================================

type InstructionCall struct {
	InstructionBase

	funcName string
	f        *Function
}

func NewInstructionCall(inst *ssa.Call) *InstructionCall {
	// TODO: handle dynamic calls
	name := ""
	if callee := inst.Common().StaticCallee(); callee != nil {
		name = callee.String()
	}

	return &InstructionCall{InstructionBase: InstructionBase{class: Ic_call},
		funcName: name}
}

func (this *InstructionCall) GetFunc(data *Data) *Function {
	if this.f != nil {
		return this.f
	}

	for _, fu := range data.funcs {
		if this.funcName == fu.name {
			this.f = fu
			return fu
		}
	}

	return nil
}

// ================================================================
// MARK: InstructionChangeInterface
// ================================================================

type InstructionChangeInterface struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionChangeType
// ================================================================

type InstructionChangeType struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionConst
// ================================================================

type InstructionConst struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionConvert
// ================================================================

type InstructionConvert struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionDebugRef
// ================================================================

type InstructionDebugRef struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionDefer
// ================================================================

type InstructionDefer struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionExtract
// ================================================================

type InstructionExtract struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionField
// ================================================================

type InstructionField struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionFieldAddr
// ================================================================

type InstructionFieldAddr struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionFreeVar
// ================================================================

type InstructionFreeVar struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionFunction
// ================================================================

type InstructionFunction struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionGo
// ================================================================

type InstructionGo struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionIf
// ================================================================

type InstructionIf struct {
	InstructionBase

	case_if   int
	case_else int
}

func newInstructionIf(inst *ssa.If) *InstructionIf {
	i := inst.Block().Succs[0].Index
	e := inst.Block().Succs[1].Index
	return &InstructionIf{InstructionBase: InstructionBase{class: Ic_jump}, case_if: i, case_else: e}
}

func (this *InstructionIf) If() int {
	return this.case_if
}

func (this *InstructionIf) Else() int {
	return this.case_else
}

// ================================================================
// MARK: InstructionIndex
// ================================================================

type InstructionIndex struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionIndexAddr
// ================================================================

type InstructionIndexAddr struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionJump
// ================================================================

type InstructionJump struct {
	InstructionBase

	to int
}

func newInstructionJump(inst *ssa.Jump) *InstructionJump {
	return &InstructionJump{InstructionBase: InstructionBase{class: Ic_jump}, to: inst.Block().Succs[0].Index}
}

func (this *InstructionJump) To() int {
	return this.to
}

// ================================================================
// MARK: InstructionLookup
// ================================================================

type InstructionLookup struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeChan
// ================================================================

type InstructionMakeChan struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeClosure
// ================================================================

type InstructionMakeClosure struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeInterface
// ================================================================

type InstructionMakeInterface struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeMap
// ================================================================

type InstructionMakeMap struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeSlice
// ================================================================

type InstructionMakeSlice struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMapUpdate
// ================================================================

type InstructionMapUpdate struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMultiConvert
// ================================================================

type InstructionMultiConvert struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionNext
// ================================================================

type InstructionNext struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionPanic
// ================================================================

type InstructionPanic struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionParameter
// ================================================================

type InstructionParameter struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionPhi
// ================================================================

type InstructionPhi struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionRange
// ================================================================

type InstructionRange struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionReturn
// ================================================================

type InstructionReturn struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionRunDefers
// ================================================================

type InstructionRunDefers struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSelect
// ================================================================

type InstructionSelect struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSend
// ================================================================

type InstructionSend struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSlice
// ================================================================

type InstructionSlice struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSliceToArrayPointer
// ================================================================

type InstructionSliceToArrayPointer struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionStore
// ================================================================

type InstructionStore struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionTypeAssert
// ================================================================

type InstructionTypeAssert struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionUnOp
// ================================================================

type InstructionUnOp struct {
	InstructionBase
}
