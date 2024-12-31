package main

import (
	"log"

	"github.com/kirill/deriv-teletrader/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
