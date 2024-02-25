package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

func parseJsonFile(filePath string) ConnectionList {
	if errors.Is(err, os.ErrNotExist) {
		log.Fatal("File Not Found Error!")
	}

	var jsonConnectionList []connection
	var jsonFile *os.File

	jsonFile, err = os.Open(filePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Can not open file(%s) ", filePath))
	}

	logChan <- logMsg{
		level: 1,
		msg:   fmt.Sprintf("File(%s) successfully opened", filePath),
	}

	decoder := json.NewDecoder(jsonFile)

	if err = decoder.Decode(&jsonConnectionList); err != nil {
		log.Fatal("Error:", err)
	}

	if verbose {
		for _, user := range jsonConnectionList {
			logChan <- logMsg{
				level: 1,
				msg:   fmt.Sprintf("User(%s) added to list", user.GetObj()),
			}
		}
	}

	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logChan <- logMsg{
				level: 3,
				msg:   fmt.Sprintf("Sip users file(%s) closing failed", filePath),
			}
		}
	}(jsonFile)

	return jsonConnectionList
}
