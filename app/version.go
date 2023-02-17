package version

import (
	"fmt"
)

var Version string

func main() {

	Version = "v0.3.1"

	fmt.Println("Version:\t", Version)

}
