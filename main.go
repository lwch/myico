package main

import (
	"fmt"
	"os"

	"github.com/lwch/myico/convert"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s <image-file-dir> <output-file-dir>\n", os.Args[0])
		return
	}
	assert(convert.Generate(os.Args[1], os.Args[2]))
}
