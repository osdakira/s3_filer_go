package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type View struct {
	viewModel *ViewModel

	app         *tview.Application
	table       *tview.Table
	filterField *tview.InputField
	pathField   *tview.InputField
}

func NewView(viewmodel *ViewModel) *View {
	view := new(View)
	view.viewModel = viewmodel
	view.app = tview.NewApplication()

	view.table = newTable()
	view.filterField = newFilterField()
	view.pathField = newPathField()

	view.SetTableToSetSelectedFunc()
	view.SetTableToSetInputCapture()
	// view.setInputCaptureOnApp()

	return view
}

func newTable() *tview.Table {
	t := tview.NewTable()
	t.SetBorders(true).SetSelectable(true, false) // rows, columns
	return t
}

func newFilterField() *tview.InputField {
	f := tview.NewInputField()
	f.SetLabel("filter: ").SetFieldBackgroundColor(tcell.ColorBlack).SetBorder(true)
	return f
}

func newPathField() *tview.InputField {
	f := tview.NewInputField()
	f.SetLabel("path: ").SetFieldBackgroundColor(tcell.ColorBlack).SetBorder(true)
	return f
}

func (self *View) Run() error {
	self.update()
	layout := self.makeLayout()
	return self.app.SetRoot(layout, true).Run()
}

func (self *View) makeLayout() tview.Primitive {
	head := tview.NewGrid()
	head.SetSize(1, 4, 0, 0).
		AddItem(self.pathField, 0, 0, 1, 3, 0, 0, false).
		AddItem(self.filterField, 0, 3, 1, 1, 0, 0, false)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).
		AddItem(head, 3, 0, false).
		AddItem(self.table, 0, 1, true)

	return flex
}

func (self *View) SetTableToSetInputCapture() {
	self.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlP: // previous line
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case tcell.KeyCtrlN: // next line
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case tcell.KeyCtrlU:
			self.viewModel.CurrentNode = self.viewModel.GetParent()
			self.update()
			return nil
		case tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2:
			text := self.filterField.GetText()
			if len(text) > 1 {
				text = text[:len(text)-1]
			} else {
				text = ""
			}
			self.filter(text)
			return nil
		case tcell.KeyRune:
			text := self.filterField.GetText() + string(event.Rune())
			self.filter(text)
			return nil
		default:
			return event
		}
	})
}

func (self *View) SetTableToSetSelectedFunc() {
	self.table.SetSelectedFunc(func(row, column int) {
		node := self.viewModel.FilteredNodes[row]
		log.Println("row", row, ", node", node)

		if !node.IsLeaf() {
			self.viewModel.CurrentNode = node
			self.update()
		}
	})
}

func (self *View) filter(text string) {
	self.filterField.SetText(text)
	nodes := self.viewModel.Filter(text)
	self.updateTable(nodes)
}

func (self *View) update() {
	node := self.viewModel.CurrentNode

	nodes := self.viewModel.UpdateCurrentNode(node)
	self.updateTable(nodes)

	self.filterField.SetText("")
	self.pathField.SetText(fmt.Sprintf("%s/%s", node.Bucket, node.Prefix))
}

func (self *View) updateTable(nodes []Node) {
	self.table.Clear()

	for r, obj := range nodes {
		values := []string{obj.Timestamp, obj.Name, obj.Size, obj.StorageClass}
		for c, val := range values {
			self.table.SetCell(r, c, tview.NewTableCell(val))
		}
	}

	self.table.ScrollToBeginning()
}

// func setInputCaptureOnApp(app *tview.Application, table *tview.Table, inputField *tview.InputField) {
// 	widgets := []tview.Primitive{table, inputField}
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		switch event.Key() {
// 		case tcell.KeyTab:
// 			current := app.GetFocus()
// 			for i, x := range widgets {
// 				if x == current {
// 					if i+1 == len(widgets) {
// 						app.SetFocus(widgets[0])
// 						break
// 					} else {
// 						app.SetFocus(widgets[i+1])
// 						break
// 					}
// 				}
// 			}
// 			return nil
// 		case tcell.KeyRune:
// 			switch event.Rune() {
// 			case '/':
// 				app.SetFocus(inputField)
// 				return nil
// 			case 'q':
// 				app.Stop()
// 				return nil
// 			}
// 		}
//
// 		return event
// 	})
// }
