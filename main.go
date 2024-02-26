package main

import (
	"sync"
)

var (
	err      error
	logChan  = make(chan logMsg)
	flags    flagType
	clientIP = getClientIP()
)

func main() {
	go logger(logChan)
	defer close(logChan)

	getFlags(&flags)

	var connectionList []Connection
	if flags.filePath == "" {
		c := Connection{
			Username: flags.username,
			Password: flags.password,
		}
		c.setResult(21)
		connectionList = append(connectionList, c)
	} else {
		connectionList = parseJsonFile(flags.filePath)
	}
	var wg sync.WaitGroup
	for _, c := range connectionList {
		validConnectionCh := validateConnectionCh(c)
		sendToStrategy(validConnectionCh, &wg)
	}
}
