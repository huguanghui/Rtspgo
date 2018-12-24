package main

import (
	"log"

	rtspgo "github.com/huguanghui/Rtspgo"
)

func main() {
	log.Fatal(rtspgo.ListenAndServe(":554"))
}
