//
// File: channel.go
// Brief: Constraints for channels
//
// Created: 2025-07-14
//
// License: BSD-3-Clause

package constraints

import "goCR/trace"

// AddChannel creates constraints for unbuffered channels
// If creates a positive constraint between the send and corresponding recv
//
// Parameter:
//   - elem *trace.ElementChannel: the element
func AddChannel(elem *trace.ElementChannel) {
	// For now, only create constraints for unbuffered channels
	// TODO: check out buffered channels
	if elem.IsBuffered() {
		return
	}

	if elem.GetOpC() == trace.RecvOp {
		p := elem.GetPartner()
		AddConstraint(true, p, elem)
	}
}
