package main

import (
	"fmt"
	"strings"
)

type Node struct {
	Name         string
	Bucket       string
	Timestamp    string
	Prefix       string
	StorageClass string
	Size         string
}

func (x Node) IsBucket() bool {
	return x.Name == x.Bucket
}

func (x Node) IsPrefix() bool {
	return strings.HasSuffix(x.Prefix, x.Name)
}

func (x Node) IsRoot() bool {
	return x.Bucket == ""
}

func (x Node) IsEdge() bool {
	return x.IsBucket() || x.IsPrefix()
}

func (x Node) IsLeaf() bool {
	return !x.IsRoot() && !x.IsEdge()
}

func (x Node) GetParent() Node {
	if x.IsBucket() {
		return Node{Bucket: ""} // make root node
	} else {
		paths := strings.Split(x.Prefix, "/") // Prefix has Trailing Slash. Last item is blank. "a/" => ["a", ""]
		if len(paths) == 2 {                  // IsBucket
			return Node{
				Bucket: x.Bucket,
				Name:   x.Bucket,
			}
		} else {
			parentPrefix := strings.Join(paths[:len(paths)-2], "/") + "/"
			name := paths[len(paths)-3] + "/"
			return Node{
				Bucket:    x.Bucket,
				Name:      name,
				Prefix:    parentPrefix,
				Timestamp: "                    ", // dummy string to keep width
			}
		}
	}
}

func (x Node) GetS3Path() string {
	return fmt.Sprintf("s3://%s/%s%s", x.Bucket, x.Prefix, x.Name)
}
