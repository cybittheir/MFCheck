package version

import (
	"fmt"
)

var Version string

func main() {

	Version = "v0.0.1"

	fmt.Println("Version:\t", Version)

}
