package main

import (
	"fmt"
	"log"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] main: %v", r)
		}
	}()

	logFile, err := initLogger()
	if err != nil {
		fmt.Println("logging init error:", err)
	} else {
		defer logFile.Close()
	}

	BuildAndRun()
}