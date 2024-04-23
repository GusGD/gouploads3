package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	i := 0
	for time.Second == time.Duration(30) {
		file, err := os.Create(fmt.Sprintf("./tmp/file-%d.txt", i))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		file.WriteString("hello world\n")
		i++
	}
}
