package main

import (
	"log"
	"path"
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
	return x.Name == ""
}

func (x Node) IsEdge() bool {
	return x.IsBucket() || x.IsPrefix()
}

func (x Node) IsLeaf() bool {
	return !x.IsRoot() && !x.IsEdge()
}

func (x Node) GetParent() Node {
	log.Println("x.Prefix", x.Prefix)

	parentPrefix := path.Dir(strings.TrimSuffix(x.Prefix, "/")) + "/"
	name := path.Base(parentPrefix) + "/"
	if parentPrefix == "./" {
		parentPrefix = ""
	}
	log.Println("parentPrefix", parentPrefix)

	return Node{
		Bucket:    x.Bucket,
		Name:      name,
		Prefix:    parentPrefix,
		Timestamp: "                    ", // dummy string to keep width
	}
}
