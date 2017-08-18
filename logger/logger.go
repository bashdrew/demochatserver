package logger

import (
	"log"
	"os"
	"strings"
)

var logFile *os.File
var err error

func TrimInput(input string) (result string) {
	result = strings.Replace(input, "\n", "", -1)
	result = strings.Replace(result, "\r", "", -1)

	return
}

func InitializeLogFile(file string) *os.File {
	logFile, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(logFile)

	return logFile
}

func PrintLog(msg string) {
	log.Println(TrimInput(msg))
}
