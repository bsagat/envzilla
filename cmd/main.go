package main

import (
	"envzilla"
	"log"
)

func main() {
	err := envzilla.Loader("configs/.env")
	if err != nil {
		log.Fatal("WTF : ", err)
	}
}
