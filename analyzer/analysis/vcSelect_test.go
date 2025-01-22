package analysis

import (
	"analyzer/clock"
	"reflect"
	"testing"
)

func TestSelectSelect(t *testing.T) {
	ClearTrace()
	casesSender := []TraceElementChannel{
		{
			routine: 1,
			tPre:    5,
			tPost:   6,
			id:      12,
			opC:     SendOp,
			cl:      false,
			oID:     456,
			qSize:   0,
			pos:     "testfile:999",
			sel:     nil,
			partner: nil,
			vc:      clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		},
	}

	sender := TraceElementSelect{
		routine:         1,
		tPre:            5,
		tPost:           6,
		id:              123,
		cases:           casesSender,
		chosenCase:      casesSender[0],
		chosenIndex:     0,
		containsDefault: false,
		chosenDefault:   false,
		pos:             "testfile:999",
		vc:              clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
	}

	casesRecv := []TraceElementChannel{
		{
			routine: 2,
			tPre:    4,
			tPost:   7,
			id:      12,
			opC:     RecvOp,
			cl:      false,
			oID:     456,
			qSize:   0,
			pos:     "testfile:888",
			sel:     nil,
			vc:      clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		},
	}

	recv := TraceElementSelect{
		routine:         2,
		tPre:            7,
		tPost:           8,
		id:              124,
		cases:           casesRecv,
		chosenCase:      casesRecv[0],
		chosenIndex:     0,
		containsDefault: false,
		chosenDefault:   false,
		pos:             "testfile:999",
		vc:              clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
	}

	traces[sender.routine] = append(traces[sender.routine], &sender)
	traces[recv.routine] = append(traces[recv.routine], &recv)

	currentVCHb = map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	RunAnalysis(false, false, make(map[string]bool), false)

	// remember that runAnalysis will increase the counter on [1][1] by 1
	expectedVcs := map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 7, 2: 9, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 6, 2: 10, 3: 7}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	if !reflect.DeepEqual(currentVCHb, expectedVcs) {
		t.Errorf("Incorrect vc. Expected %v. Got %v.", expectedVcs, currentVCHb)
	}
}

func TestChanSelect(t *testing.T) {
	ClearTrace()

	sender := TraceElementChannel{
		routine: 1,
		tPre:    5,
		tPost:   6,
		id:      12,
		opC:     SendOp,
		cl:      false,
		oID:     456,
		qSize:   0,
		pos:     "testfile:999",
		sel:     nil,
		partner: nil,
		vc:      clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
	}

	casesRecv := []TraceElementChannel{
		{
			routine: 2,
			tPre:    4,
			tPost:   7,
			id:      12,
			opC:     RecvOp,
			cl:      false,
			oID:     456,
			qSize:   0,
			pos:     "testfile:888",
			sel:     nil,
			vc:      clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		},
	}

	recv := TraceElementSelect{
		routine:         2,
		tPre:            7,
		tPost:           8,
		id:              124,
		cases:           casesRecv,
		chosenCase:      casesRecv[0],
		chosenIndex:     0,
		containsDefault: false,
		chosenDefault:   false,
		pos:             "testfile:999",
		vc:              clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
	}

	traces[sender.routine] = append(traces[sender.routine], &sender)
	traces[recv.routine] = append(traces[recv.routine], &recv)

	currentVCHb = map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	RunAnalysis(false, false, make(map[string]bool), false)

	// remember that runAnalysis will increase the counter on [1][1] by 1
	expectedVcs := map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 7, 2: 9, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 6, 2: 10, 3: 7}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	if !reflect.DeepEqual(currentVCHb, expectedVcs) {
		t.Errorf("Incorrect vc. Expected %v. Got %v.", expectedVcs, currentVCHb)
	}
}

func TestSelectChan(t *testing.T) {
	ClearTrace()
	casesSender := []TraceElementChannel{
		{
			routine: 1,
			tPre:    5,
			tPost:   6,
			id:      12,
			opC:     SendOp,
			cl:      false,
			oID:     456,
			qSize:   0,
			pos:     "testfile:999",
			sel:     nil,
			partner: nil,
			vc:      clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		},
	}

	sender := TraceElementSelect{
		routine:         1,
		tPre:            5,
		tPost:           6,
		id:              123,
		cases:           casesSender,
		chosenCase:      casesSender[0],
		chosenIndex:     0,
		containsDefault: false,
		chosenDefault:   false,
		pos:             "testfile:999",
		vc:              clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
	}

	recv := TraceElementChannel{
		routine: 2,
		tPre:    4,
		tPost:   7,
		id:      12,
		opC:     RecvOp,
		cl:      false,
		oID:     456,
		qSize:   0,
		pos:     "testfile:888",
		sel:     nil,
		vc:      clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
	}

	traces[sender.routine] = append(traces[sender.routine], &sender)
	traces[recv.routine] = append(traces[recv.routine], &recv)

	currentVCHb = map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	RunAnalysis(false, false, make(map[string]bool), false)

	// remember that runAnalysis will increase the counter on [1][1] by 1
	expectedVcs := map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 7, 2: 9, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 6, 2: 10, 3: 7}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	if !reflect.DeepEqual(currentVCHb, expectedVcs) {
		t.Errorf("Incorrect vc. Expected %v. Got %v.", expectedVcs, currentVCHb)
	}
}

func TestDefault(t *testing.T) {
	ClearTrace()
	casesSender := []TraceElementChannel{
		{
			routine: 1,
			tPre:    5,
			tPost:   6,
			id:      12,
			opC:     SendOp,
			cl:      false,
			oID:     456,
			qSize:   0,
			pos:     "testfile:999",
			sel:     nil,
			partner: nil,
			vc:      clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		},
	}

	sel := TraceElementSelect{
		routine:         1,
		tPre:            5,
		tPost:           6,
		id:              123,
		cases:           casesSender,
		chosenCase:      TraceElementChannel{},
		chosenIndex:     -1,
		containsDefault: true,
		chosenDefault:   true,
		pos:             "testfile:999",
		vc:              clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
	}

	traces[sel.routine] = append(traces[sel.routine], &sel)

	currentVCHb = map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	RunAnalysis(false, false, make(map[string]bool), false)

	// remember that runAnalysis will increase the counter on [1][1] by 1
	expectedVcs := map[int]clock.VectorClock{
		1: clock.NewVectorClockSet(3, map[int]int{1: 7, 2: 6, 3: 7}),
		2: clock.NewVectorClockSet(3, map[int]int{1: 2, 2: 9, 3: 4}),
		3: clock.NewVectorClockSet(3, map[int]int{1: 5, 2: 6, 3: 9}),
	}

	if !reflect.DeepEqual(currentVCHb, expectedVcs) {
		t.Errorf("Incorrect vc. Expected %v. Got %v.", expectedVcs, currentVCHb)
	}
}
