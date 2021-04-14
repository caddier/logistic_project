package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func WhereAmI(depthList ...int) string {
	var depth int
	if depthList == nil {
		depth = 2
	} else {
		depth = depthList[0]
	}
	_, file, line, _ := runtime.Caller(depth)
	file = path.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
}

func initLog(logfile string) {
	log.SetFlags(log.LstdFlags)
	file, _ := os.Create(logfile)
	log.SetOutput(file)
}

func LogInfo(format string, args ...interface{}) {
	newFmt := fmt.Sprintf("%s.%d[I]%s\n", WhereAmI(), getGID(), format)
	log.Printf(newFmt, args...)
}

func LogError(format string, args ...interface{}) {
	newFmt := fmt.Sprintf("%s.%d[E]%s\n", WhereAmI(), getGID(), format)
	log.Printf(newFmt, args...)
}

func Gracefull() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Block until a signal is received.
	<-c
}
