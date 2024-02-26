package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
)

type flagType struct {
	verbose   bool
	strategy  string
	protocol  string
	algorithm string
	filePath  string
	username  string
	password  string
	uri       string
	port      int
	workers   int
}

var (
	err         error
	logChan     = make(chan logMsg)
	flags       flagType
	Connections []ConnectionTools
	clientIP    = getClientIP()
)

func getFlags(f *flagType) {
	argsError := make([]string, 0, 2)

	flag.BoolVar(&f.verbose, "v", false, "Verbose Mode")
	flag.IntVar(&f.workers, "w", 1, "Workers Count")
	flag.StringVar(&f.strategy, "s", "register", "SIP Packet strategy(register, invite)")
	flag.StringVar(&f.protocol, "p", "", "SIP Server Auth Protocol")
	flag.StringVar(&f.algorithm, "a", "", "SIP Server Auth Algorithm(MD5 or MD5-SESS)")
	flag.StringVar(&f.filePath, "f", "", "SIP Users File Path")
	flag.StringVar(&f.username, "username", "", "SIP username")
	flag.StringVar(&f.password, "password", "", "SIP Password")
	flag.StringVar(&f.uri, "uri", "", "SIP Server Address(uri)")
	flag.IntVar(&f.port, "port", 5060, "SIP Server Port")

	flag.Parse()

	f.protocol = strings.ToLower(f.protocol)
	if f.protocol != "tcp" && f.protocol != "udp" {
		f.protocol = "tcp"
		logChan <- logMsg{
			level: 2,
			msg:   "protocol not set or valid. set to default to tcp",
		}
	}

	f.strategy = strings.ToUpper(f.strategy)
	if !strings.Contains(strings.Join([]string{"REGISTER", "INVITE"}, ","), f.strategy) {
		f.strategy = "REGISTER"
		logChan <- logMsg{
			level: 2,
			msg:   "mode not set. set to default(REGISTER)",
		}
	}

	if f.algorithm != "MD5" && f.algorithm != "MD5-SESS" {
		f.algorithm = "MD5"
	}

	if f.uri == "" {
		argsError = append(argsError, "SIP Server Address(uri) is required!")
	}

	if f.filePath == "" {
		if f.username == "" {
			argsError = append(argsError, "SIP username is required!")
		}
		if f.password == "" {
			const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
			passwordBytes := make([]byte, 8)
			for i := range passwordBytes {
				passwordBytes[i] = letterBytes[rand.Intn(len(letterBytes))]
			}
			f.password = string(passwordBytes)
			logChan <- logMsg{
				level: 2,
				msg:   "SIP password is required! Generate random password in progress",
			}
			logChan <- logMsg{
				level: 1,
				msg:   fmt.Sprintf("SIP password is: %s", f.password),
			}
		}
	} else {
		if !strings.HasPrefix(f.filePath, "/") {
			var pwd string
			pwd, _ = os.Getwd()
			f.filePath = fmt.Sprintf("%s/%s", pwd, f.filePath)
		}
	}
	if len(argsError) > 0 {
		note := []string{"\nRequired args:", strings.Join(argsError[:], "\n")}
		log.Fatal(strings.Join(note[:], "\n"))
	}
}

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
		connectionList = append(connectionList, c)
	} else {
		connectionList = parseJsonFile(flags.filePath)
	}

	validConnectionCh := validateConnectionCh(connectionList)
	progressConnectionCh := connectionStatus(validConnectionCh)
	var wg sync.WaitGroup
	__sendToStrategy(progressConnectionCh, &wg)

}
