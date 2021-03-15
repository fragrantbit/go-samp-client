package main

import (
	"log"
	"time"
	"fmt"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
    return fmt.Print(time.Now().Format("[15:04:05] ") + string(bytes))
}

func Init() {
	log.SetFlags(0)
    log.SetOutput(new(logWriter))
}