package main

import (
	"log"

	"github.com/kirill/deriv-teletrader/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
