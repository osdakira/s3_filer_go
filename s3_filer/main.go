package main

import (
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/cli/safeexec"
)

func main() {
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

func init() {
	if os.Getenv("LC_CTYPE") != "en_US.UTF-8" && runtime.GOOS != "windows" {
		err := os.Setenv("LC_CTYPE", "en_US.UTF-8")
		if err != nil {
			panic(err)
		}
		env := os.Environ()
		argv0, err := safeexec.LookPath(os.Args[0])
		if err != nil {
			panic(err)
		}
		os.Args[0] = argv0
		/* #nosec G204 */
		if err := syscall.Exec(argv0, os.Args, env); err != nil {
			panic(err)
		}
	}
}
