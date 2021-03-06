package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const SAVE_PATH = "/tmp/s3_filer_go_84EB04C0"

type ViewModel struct {
	Buckets       []Node
	Nodes         []Node
	FilteredNodes []Node
	CurrentNode   Node
	Client        *s3.Client
	downloader    *manager.Downloader
}

func NewViewModel() *ViewModel {
	viewModel := new(ViewModel)
	viewModel.Client = buildClient()
	viewModel.downloader = buildDownloader(viewModel.Client)
	viewModel.Buckets = GetAllBuckets(viewModel.Client)

	node, err := viewModel.Load()
	if err != nil {
		node = viewModel.GetRootLikeNode()
	}

	viewModel.UpdateCurrentNode(node)
	return viewModel
}

func buildClient() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return s3.NewFromConfig(cfg)
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
			Timestamp: x.CreationDate.Format(time.RFC3339),
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (self *ViewModel) GetRootLikeNode() Node {
	return Node{Bucket: ""}
}

func (self *ViewModel) UpdateCurrentNode(node Node) []Node {
	if node.IsRoot() {
		self.Nodes = self.Buckets
	} else if node.IsEdge() {
		self.Nodes = self.fetchChildren(node)
	} else {
		return nil
	}

	self.FilteredNodes = self.Nodes
	self.CurrentNode = node
	return self.FilteredNodes
}

func (self *ViewModel) fetchChildren(node Node) []Node {
	output, err := self.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
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
			Timestamp: "                    ", // dummy string to keep width
		}
		nodes = append(nodes, node)
	}

	for _, x := range output.Contents {
		key := aws.ToString(x.Key)
		node := Node{
			Bucket:       node.Bucket,
			Name:         strings.Replace(key, node.Prefix, "", 1),
			Prefix:       node.Prefix,
			Timestamp:    x.LastModified.Format(time.RFC3339),
			Size:         strconv.FormatInt(x.Size, 10),
			StorageClass: string(x.StorageClass),
		}
		nodes = append(nodes, node)
	}

	return nodes
}

func (self *ViewModel) GetParent() Node {
	return self.CurrentNode.GetParent()
}

func (self *ViewModel) Filter(text string) []Node {
	if text == "" {
		self.FilteredNodes = self.Nodes
	} else {
		searchReg := strings.Join(strings.Split(text, ""), "+")
		r := regexp.MustCompile(searchReg)

		var newNodes []Node
		for _, x := range self.Nodes {
			if r.MatchString(x.Name) {
				newNodes = append(newNodes, x)
			}
		}
		self.FilteredNodes = newNodes
	}
	return self.FilteredNodes
}

func buildDownloader(client *s3.Client) *manager.Downloader {
	return manager.NewDownloader(client)
}

func (self *ViewModel) Download(node Node) error {
	f, err := os.Create(node.Name)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = self.downloader.Download(context.Background(), f, &s3.GetObjectInput{
		Bucket: aws.String(node.Bucket),
		Key:    aws.String(node.Prefix + node.Name),
	})
	return err
}

func (self *ViewModel) FetchFirst(node Node) (string, error) {
	buf, err := self.GetObjectFromS3(node, "bytes=0-500")
	if err != nil {
		return string(buf), err
	}

	ftype, err := GuessFileType(buf)
	if err != nil {
		return ftype, err
	}

	switch {
	case strings.Contains(ftype, "ASCII text"):
		return string(buf), nil
	case strings.Contains(ftype, "gzip"):
		return ReadGzip(buf)
	case strings.Contains(ftype, "Apache Parquet"):
		buf2, err := self.GetObjectFromS3(node, "")
		if err != nil {
			return "", err
		}
		return ReadParquet(buf2)
	default:
		return "Binary file not shown", nil
	}

	return string(buf), nil
}

func (self *ViewModel) GetObjectFromS3(node Node, byteRange string) ([]byte, error) {
	input := s3.GetObjectInput{
		Bucket: aws.String(node.Bucket),
		Key:    aws.String(node.Prefix + node.Name),
	}
	if byteRange != "" {
		input.Range = aws.String(byteRange)
	}

	resp, err := self.Client.GetObject(context.TODO(), &input)
	if err != nil {
		Debug("GetObject err", err)
		return []byte{}, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		Debug("GetObject ReadAll err", err)
		return []byte{}, err
	}

	return buf, nil
}

func (self *ViewModel) Save() error {
	json, err := json.Marshal(self.CurrentNode)
	if err != nil {
		return err
	}

	err = os.WriteFile(SAVE_PATH, json, 0644)
	return err
}

func (self *ViewModel) Load() (Node, error) {
	var node Node

	raw, err := os.ReadFile(SAVE_PATH)
	if err != nil {
		return node, err
	}

	err = json.Unmarshal(raw, &node)
	if err != nil {
		return node, err
	}

	return node, nil
}
