package logging

import (
	"log"
)

var debug bool = false

func EnableDebugMode() {
	debug = true
}

func Println(args ...interface{}) {
	if debug {
		log.Println(args...)
	}
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}
