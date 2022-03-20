package main

import (
	"log"
	"os"
)

func main() {
	if err := os.Setenv("LC_CTYPE", "en_US.UTF-8"); err != nil {
		panic(err)
	}

	f := setLogger("debug.log")
	defer f.Close()

	viewModel := NewViewModel()
	view := NewView(viewModel)

	if err := view.Run(); err != nil {
		panic(err)
	}
}

func setLogger(path string) *os.File {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	return f
}
