package greeting

import (
	"fmt"
	"syscall"
)

var (
	appname      string
	version      string
	build        string
	getTickCount      = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")
	silent       bool = false // display all messages
)

func Greeting() {

	fmt.Println(appname, version, "build", build)
	fmt.Println("Simple agent for checking links to devices and running applications on remote PC")
	fmt.Println("https://github.com/cybittheir/MFCheck")

}
