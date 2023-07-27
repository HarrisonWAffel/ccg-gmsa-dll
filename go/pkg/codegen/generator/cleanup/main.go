package main

import (
	"os"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := os.RemoveAll("./go/pkg/generated"); err != nil {
		return err
	}
	return nil
}