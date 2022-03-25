package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type View struct {
	viewModel *ViewModel

	app         *tview.Application
	table       *tview.Table
	filterField *tview.InputField
	pathField   *tview.InputField
	statusBar   *tview.InputField
	modal       *tview.TextView
	pages       *tview.Pages

	humanReadable bool
}

func NewView(viewmodel *ViewModel) *View {
	view := new(View)
	view.viewModel = viewmodel
	view.app = tview.NewApplication()

	view.table = newTable()
	view.filterField = newFilterField()
	view.pathField = newPathField()
	view.statusBar = newStatusBar()
	view.modal = newModal(view)

	view.SetTableToSetInputCapture()
	view.setInputCaptureOnApp()

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

func newStatusBar() *tview.InputField {
	f := tview.NewInputField()
	f.SetFieldBackgroundColor(tcell.ColorBlack)
	return f
}

func newModal(view *View) *tview.TextView {
	modal := tview.NewTextView()
	modal.SetBorder(true).SetTitle("View: first 500 bytes")
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyEsc:
			view.pages.SwitchToPage("base")
			return nil
		default:
			return event
		}
	})
	return modal
}

func (self *View) Run() error {
	self.update()
	layout := self.makeLayout()
	err := self.app.SetRoot(layout, true).Run()
	if err != nil {
		self.viewModel.Save()
	}
	return err
}

func (self *View) makeLayout() tview.Primitive {
	head := tview.NewGrid()
	head.SetSize(1, 4, 0, 0).
		AddItem(self.pathField, 0, 0, 1, 3, 0, 0, false).
		AddItem(self.filterField, 0, 3, 1, 1, 0, 0, false)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).
		AddItem(head, 3, 0, false).
		AddItem(self.table, 0, 1, true).
		AddItem(self.statusBar, 1, 0, false)

	self.pages = tview.NewPages().
		AddPage("modal", self.modal, true, true).
		AddPage("base", flex, true, true)
		// AddPage(name string, item Primitive, resize, visible bool)

	return self.pages
}

func (self *View) SetTableToSetInputCapture() {
	self.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		self.statusBar.SetText("")

		switch event.Key() {
		case tcell.KeyCtrlP: // previous line
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case tcell.KeyCtrlN: // next line
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case tcell.KeyEnter, tcell.KeyRight:
			self.intoSelectionNode()
			return nil
		case tcell.KeyCtrlU, tcell.KeyLeft:
			self.viewModel.CurrentNode = self.viewModel.GetParent()
			self.update()
			return nil
		case tcell.KeyCtrlD:
			self.DownloadNode()
			return nil
		case tcell.KeyCtrlH:
			self.humanReadable = !self.humanReadable
			self.filter(self.filterField.GetText())
			return nil
		case tcell.KeyDelete, tcell.KeyBackspace2:
			self.deleteFilterWord()
			return nil
		case tcell.KeyRune:
			self.addFilterWord(string(event.Rune()))
			return nil
		default:
			return event
		}
	})
}

func (self *View) DownloadNode() {
	row, _ := self.table.GetSelection()
	node := self.viewModel.FilteredNodes[row]
	message, err := self.viewModel.Download(node)
	if err != nil {
		self.statusBar.SetText(fmt.Sprintf("%v", err))
	} else {
		self.statusBar.SetText(message)
	}
}

func (self *View) intoSelectionNode() {
	row, _ := self.table.GetSelection()
	if row >= len(self.viewModel.FilteredNodes) {
		return
	}

	node := self.viewModel.FilteredNodes[row]
	if node.IsLeaf() {
		text, err := self.viewModel.FetchFirst(node)
		if err != nil {
			log.Println(err)
		}
		self.modal.SetText(text)
		self.pages.SwitchToPage("modal")
	} else {
		self.viewModel.CurrentNode = node
		self.update()
	}
}

func (self *View) addFilterWord(word string) {
	text := self.filterField.GetText() + word
	self.filter(text)
}

func (self *View) deleteFilterWord() {
	text := self.filterField.GetText()
	if len(text) > 1 {
		text = text[:len(text)-1]
	} else {
		text = ""
	}
	self.filter(text)
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
		size := obj.Size
		if self.humanReadable {
			size64, err := strconv.ParseInt(size, 10, 64)
			if err == nil {
				size = ByteCountIEC(size64)
			}
		}
		values := []string{obj.Timestamp, obj.Name, size, obj.StorageClass}
		for c, val := range values {
			self.table.SetCell(r, c, tview.NewTableCell(val))
		}
	}

	self.table.ScrollToBeginning()
}

func (self *View) setInputCaptureOnApp() {
	self.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			self.viewModel.Save()
			self.app.Stop()
		}
		return event
	})
}
