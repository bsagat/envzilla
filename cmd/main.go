package main

import (
	"fmt"
	"os"

	"github.com/bsagat/envzilla"
)

func main() {
	cfg := struct {
		Secret string `env:"secret"`
	}{}

	os.Setenv("secret", "")

	fmt.Println(envzilla.Parse(&cfg))

	fmt.Println(cfg)
}
