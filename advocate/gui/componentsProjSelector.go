// Copyright (c) 2026 Erik Kassubek
//
// File: componentsMainTestSelector.go
// Brief: Create main/test selector
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type componentProjectSelector struct {
	*fyne.Container

	w *window

	selectProj        *fyne.Container
	selectedProjLabel *widget.Label
	openProjButton    *widget.Button

	path string
}

func createProjSelector(win *window) componentProjectSelector {
	cps := componentProjectSelector{
		w: win,
	}

	cps.selectedProjLabel = widget.NewLabel("No project selected")

	cps.openProjButton = widget.NewButtonWithIcon(
		"Choose Project",
		theme.FolderOpenIcon(),
		func() {
			fileDialog := dialog.NewFolderOpen(
				func(uri fyne.ListableURI, err error) {
					if err != nil {
						win.appendOutput("Error opening folder dialog", log.ErrorLv)
						return
					}

					if uri == nil {
						return
					}

					path := uri.Path()
					cps.selectedProjLabel.SetText(filepath.Base(path))

					cps.path = path
					flags.ProgPath = path
					cps.getAllTestNames()
				},
				win.w,
			)

			fileDialog.Show()
		},
	)

	cps.Container = container.NewVBox(
		widget.NewLabel("Project:"),
		cps.openProjButton,
		cps.selectedProjLabel,
	)

	return cps
}

func (self *componentProjectSelector) getAllTestNames() {
	if self.path == "" {
		return
	}

	var testNames []string

	err := filepath.Walk(self.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}

			if strings.HasPrefix(fn.Name.Name, "Test") {
				testNames = append(testNames, fn.Name.Name)
			}
		}

		return nil
	})

	if err != nil {
		self.w.appendOutput(err.Error(), log.ErrorLv)
	}

	sort.Strings(testNames)

	self.w.mainTestSelect.setTestNames(&testNames)
}
