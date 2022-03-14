package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Node struct {
	Name         string
	Bucket       string
	IsBucket     bool
	Timestamp    string
	Prefix       string
	StorageClass string
	Size         string
	IsPrefix     bool
}

type ViewviewModel struct {
	Buckets       []Node
	Nodes         []Node
	FilteredNodes []Node
}

func main() {
	f := setLogger("debug.log")
	defer f.Close()

	client := buildClient()

	viewModel := new(ViewviewModel)
	viewModel.Buckets = GetAllBuckets(client)
	viewModel.Nodes = viewModel.Buckets
	viewModel.FilteredNodes = viewModel.Nodes

	app := tview.NewApplication()

	table := tview.NewTable()
	table.SetBorders(true).SetSelectable(true, false) // rows, columns

	updateTable(table, viewModel.FilteredNodes)

	inputField := makeFilterField(app, table, viewModel)
	pathField := makePathField()

	table.SetSelectedFunc(func(row, column int) {
		log.Println("row", row, ", node", viewModel.FilteredNodes[row])

		node := viewModel.FilteredNodes[row]
		if node.IsBucket || node.IsPrefix {
			inputField.SetText("")
			pathField.SetText(fmt.Sprintf("%s/%s", node.Bucket, node.Prefix))

			viewModel.Nodes = getObjects(client, node, table)
			viewModel.FilteredNodes = viewModel.Nodes
			updateTable(table, viewModel.FilteredNodes)
		}
	})

	setInputCaptureOnApp(app, table, inputField)

	head := tview.NewGrid()
	head.SetSize(1, 4, 0, 0).
		AddItem(pathField, 0, 0, 1, 3, 0, 0, false).
		AddItem(inputField, 0, 3, 1, 1, 0, 0, true)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).
		AddItem(head, 3, 0, true).
		AddItem(table, 0, 1, false)
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

func setInputCaptureOnApp(app *tview.Application, table *tview.Table, inputField *tview.InputField) {
	widgets := []tview.Primitive{table, inputField}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			current := app.GetFocus()
			for i, x := range widgets {
				if x == current {
					if i+1 == len(widgets) {
						app.SetFocus(widgets[0])
						break
					} else {
						app.SetFocus(widgets[i+1])
						break
					}
				}
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case '/':
				app.SetFocus(inputField)
				return nil
			case 'q':
				app.Stop()
				return nil
			}
		}

		return event
	})
}

func updateTable(table *tview.Table, nodes []Node) {
	table.Clear()
	for r, obj := range nodes {
		values := []string{obj.Timestamp, obj.Name, obj.Size, obj.StorageClass}
		for c, val := range values {
			table.SetCell(r, c, tview.NewTableCell(val))
		}
	}
	//
	// STANDARD
	// DEEP_ARCHIVE
	table.ScrollToBeginning()
}

func getObjects(client *s3.Client, node Node, table *tview.Table) []Node {
	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(node.Bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(node.Prefix),
	})
	if err != nil {
		log.Fatal(err)
	}

	var nodes []Node

	for _, x := range output.CommonPrefixes {
		prefix := aws.ToString(x.Prefix)
		node := Node{
			Bucket:    node.Bucket,
			Name:      strings.Replace(prefix, node.Prefix, "", 1),
			Prefix:    prefix,
			IsBucket:  false,
			IsPrefix:  true,
			Timestamp: "                    ", // dummy string to keep width
		}
		nodes = append(nodes, node)
	}

	for _, x := range output.Contents {
		node := Node{
			Bucket:       node.Bucket,
			Name:         strings.Replace(aws.ToString(x.Key), node.Prefix, "", 1),
			Prefix:       node.Prefix,
			IsBucket:     false,
			IsPrefix:     false,
			Timestamp:    x.LastModified.Format(time.RFC3339),
			Size:         strconv.FormatInt(x.Size, 10),
			StorageClass: string(x.StorageClass),
		}
		nodes = append(nodes, node)
	}

	return nodes
}

func GetAllBuckets(client *s3.Client) []Node {
	result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}

	var nodes []Node
	for _, x := range result.Buckets {
		node := Node{
			Bucket:    aws.ToString(x.Name),
			Name:      aws.ToString(x.Name),
			IsBucket:  true,
			IsPrefix:  false,
			Timestamp: x.CreationDate.Format(time.RFC3339),
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func setLogger(path string) *os.File {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	return f
}

func buildClient() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return s3.NewFromConfig(cfg)
}

func makeFilterField(app *tview.Application, table *tview.Table, viewModel *ViewviewModel) *tview.InputField {
	inputField := tview.NewInputField()
	inputField.SetLabel("filter: ")
	inputField.SetBorder(true)
	inputField.SetFieldBackgroundColor(tcell.ColorBlack)
	inputField.SetChangedFunc(func(text string) {
		if text == "" {
			viewModel.FilteredNodes = viewModel.Nodes
		} else {
			searchReg := strings.Join(strings.Split(text, ""), "+")
			// log.Println(searchReg)
			r := regexp.MustCompile(searchReg)

			var newNodes []Node
			for _, x := range viewModel.Nodes {
				if r.MatchString(x.Name) {
					newNodes = append(newNodes, x)
				}
			}
			viewModel.FilteredNodes = newNodes
		}
		updateTable(table, viewModel.FilteredNodes)
	})

	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			app.SetFocus(table)
			return nil
		}
		return event
	})

	return inputField
}

func makePathField() *tview.InputField {
	pathField := tview.NewInputField()
	pathField.SetLabel("path: ")
	pathField.SetBorder(true)
	pathField.SetFieldBackgroundColor(tcell.ColorBlack)
	return pathField
}
