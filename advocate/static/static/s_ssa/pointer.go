// DOES NOT WORK
// DO NOT USE

// Code to build and run points-to analysis based on pointer package
// "golang.org/x/tools/go/pointer" is depricated and does not work with modern Go versions

package s_ssa

// import (
// 	"fmt"
// 	"go/types"

// 	"golang.org/x/tools/go/pointer"
// 	"golang.org/x/tools/go/ssa"
// )

// func (self *Data) buildPointer() error {
// 	config := &pointer.Config{
// 		Mains: self.ssaMains,
// 	}

// 	isPointerLike := func(v ssa.Value) bool {
// 		t := v.Type()
// 		if t == nil {
// 			return false
// 		}

// 		switch t.Underlying().(type) {
// 		case *types.Pointer,
// 			*types.Interface,
// 			*types.Slice,
// 			*types.Map,
// 			*types.Chan,
// 			*types.Signature:
// 			return true
// 		}
// 		return false
// 	}

// 	add := func(v ssa.Value) {
// 		if v == nil {
// 			return
// 		}
// 		if isPointerLike(v) {
// 			config.AddQuery(v)
// 		}
// 	}

// 	for _, pkg := range self.ssaPkgs {
// 		for _, mem := range pkg.Members {
// 			fn, ok := mem.(*ssa.Function)
// 			if !ok || fn.Blocks == nil {
// 				continue
// 			}

// 			for _, block := range fn.Blocks {
// 				for _, instr := range block.Instrs {

// 					v, ok := instr.(ssa.Value)
// 					if !ok || v == nil {
// 						continue
// 					}

// 					if _, ok := v.Type().(*types.Tuple); ok {
// 						continue
// 					}

// 					switch v := v.(type) {

// 					case *ssa.Alloc,
// 						*ssa.FieldAddr,
// 						*ssa.IndexAddr,
// 						*ssa.Global:
// 						add(v)

// 					case *ssa.MakeSlice,
// 						*ssa.MakeMap,
// 						*ssa.MakeChan:
// 						add(v)

// 					case *ssa.Call:
// 						// only keep calls that return pointer-like results
// 						if _, ok := v.Type().(*types.Tuple); ok {
// 							continue // multiple return values (or void-like)
// 						}
// 						if isPointerLike(v) {
// 							add(v)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	result, err := pointer.Analyze(config)
// 	if err != nil {
// 		return err
// 	}

// 	self.ptrConfig = config
// 	self.ptrResult = result

// 	return nil
// }

// func (self *Data) lookup(v ssa.Value) {
// 	pts := self.ptrResult.Queries[v]

// 	fmt.Println(pts.PointsTo())
// }

// func (self *Data) mayAlias(v1, v2 ssa.Value) {
// 	p1 := self.ptrResult.Queries[v1]
// 	p2 := self.ptrResult.Queries[v2]

// 	if p1.PointsTo().Intersects(p2.PointsTo()) {
// 		fmt.Println("may alias")
// 	} else {
// 		fmt.Println("no alias")
// 	}
// }
