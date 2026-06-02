// Copyright (c) 2026 Erik Kassubek
//
// File: output.go
// Brief: Output
//
// Author: Erik Kassubek
// Created: 2026-06-02
//
// License: BSD-3-Clause

package log

type ChannelWriter struct {
	ch chan<- GuiInfo
}

func NewChannelWriter() ChannelWriter {
	return ChannelWriter{ch: guiChan}
}

func (w ChannelWriter) Write(p []byte) (n int, err error) {
	if !guiChanSet {
		return 0, nil
	}

	w.ch <- GuiInfo{Msg: string(p), Lv: OutputLv}
	return len(p), nil
}

func (w ChannelWriter) IsSet() bool {
	return guiChanSet
}
