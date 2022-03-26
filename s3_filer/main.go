package main

import (
	"os"
	"runtime"
	"syscall"

	"github.com/cli/safeexec"
)

func main() {
	viewModel := NewViewModel()
	view := NewView(viewModel)
	if err := view.Run(); err != nil {
		panic(err)
	}
}

// main の起動前に、 LC_TYPE を上書きする。 tview が壊れるから
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
