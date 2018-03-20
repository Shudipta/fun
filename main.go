package main

import (
	"fun/docker"
	"fmt"
)

func main() {
	fmt.Println(docker.GetLabels("shudipta/labels"))
}
