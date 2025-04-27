// Package logging is for logging the performance and prompting outputs of the pipelines.
package logging

import (
	"log"
	"os"

	"github.com/StoneG24/slape/pkg/vars"
)

// CreateLogFile is used to check and see if a logfile is already created.
// It then creates a logger for the log file and returns it.
func CreateLogFile() *os.File {
    logFile, err := os.OpenFile("./logs/"+vars.Logfilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("Error creating the log file")
	}

	log.SetOutput(logFile)
	return logFile
}

// CloseLogging moves the current runs logs to a trunk log file.
func CloseLogging() {

	// read all of the contents of the log file
	content, err := os.ReadFile("./logs/" + vars.Logfilename)
	if err != nil {
		log.Println("Error trying to read ", vars.Logfilename)
	}

    // open the trunk file for appending
    trunkFile, err := os.OpenFile("./logs/"+vars.Trunkfilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
    if err != nil {
        log.Println("Error creating the trunk file")
    }

	// move to the big log file
	_, err = trunkFile.Write(content)
	if err != nil {
		log.Println("Error while trying to update old logs with current run")
	}

	trunkFile.Close()
}
