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
	"fmt"
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

type componentPathSelector struct {
	*fyne.Container

	label *componentSectionLabel

	selectPath        *fyne.Container
	selectedPathLabel *widget.Label
	openPathSelButton *widget.Button

	path string
}

func createPathSelector(label string, valToSet *string) *componentPathSelector {
	cps := &componentPathSelector{}

	cps.selectedPathLabel = widget.NewLabel(fmt.Sprintf("No %s selected", strings.ToLower(label)))

	cps.openPathSelButton = widget.NewButtonWithIcon(
		"Select",
		theme.FolderOpenIcon(),
		func() {
			fileDialog := dialog.NewFolderOpen(
				func(uri fyne.ListableURI, err error) {
					if err != nil {
						win.writeErr("Error opening folder dialog")
						return
					}

					if uri == nil {
						return
					}

					path := uri.Path()
					cps.selectedPathLabel.SetText(filepath.Base(path))

					cps.path = path
					*valToSet = path
					cps.getAllTestNames()
				},
				win.w,
			)

			fileDialog.Show()
		},
	)

	cps.label = createSectionLabel(label)

	cps.Container = container.NewVBox(
		cps.label.Container,
		cps.openPathSelButton,
		cps.selectedPathLabel,
	)

	return cps
}

func (self *componentPathSelector) getAllTestNames() {
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
		win.writeErr(err.Error())
	}

	sort.Strings(testNames)

	win.settings.components.mainTestSelect.setTestNames(&testNames)
}

func (self *componentPathSelector) disable() {
	self.openPathSelButton.Disable()
}

func (self *componentPathSelector) enable() {
	self.openPathSelButton.Enable()
}
