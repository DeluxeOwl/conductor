package conductor

import (
	"os"
)

var (
	LogFilePath *string
)

func initLogFile() *os.File {
	filePath := os.DevNull
	if LogFilePath != nil {
		filePath = *LogFilePath
	}

	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	return f
}
