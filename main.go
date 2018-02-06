package main

import(
	"fmt"
	"fun/docker"
)

func main() {
	fmt.Println(docker.GetLabels("tigerworks/labels:latest"))
}
