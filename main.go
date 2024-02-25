package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type flagType struct {
	verbose           bool
	mode              string
	protocol          string
	algorithm         string
	filePath          string
	sipSimpleUsername string
	sipSimplePassword string
	sipSimpleUri      string
	sipSimplePort     int
	workers           int
}

type ConnectionList []Connection

var (
	err            error
	logChan        = make(chan logMsg)
	connectionList ConnectionList
	verbose        bool
	workers        int
)

func getFlags(f *flagType) {
	flag.BoolVar(&f.verbose, "v", false, "Verbose Mode")
	flag.IntVar(&f.workers, "workers", 1, "Verbose Mode")
	//flag.StringVar(&f.mode, "mode", "register", "SIP Packet Type(register=1, invite=2)")
	flag.StringVar(&f.protocol, "protocol", "tcp", "SIP Server Auth Protocol")
	flag.StringVar(&f.algorithm, "algorithm", "md5", "SIP Server Auth Algorithm(MD5 or MD5-SESS)")
	flag.StringVar(&f.filePath, "file", "", "SIP Users File Path")
	flag.StringVar(&f.sipSimpleUsername, "username", "", "SIP username")
	flag.StringVar(&f.sipSimplePassword, "password", "", "SIP Password")
	flag.StringVar(&f.sipSimpleUri, "uri", "", "SIP Server Address(uri)")
	flag.IntVar(&f.sipSimplePort, "port", 5060, "SIP Server Port")
	flag.Parse()

	verbose = f.verbose
	workers = f.workers

	f.protocol = strings.ToLower(f.protocol)
	if f.protocol != "tcp" && f.protocol != "udp" {
		f.protocol = "tcp"
		logChan <- logMsg{
			level: 3,
			msg:   "protocol not valid. set to default(tcp)",
		}
	}

	f.mode = strings.ToUpper(f.mode)
	if !strings.Contains(strings.Join([]string{"REGISTER", "INVITE"}, ","), f.mode) {
		f.mode = "REGISTER"
		logChan <- logMsg{
			level: 3,
			msg:   "mode not set. set to default(REGISTER)",
		}
	}

	if f.algorithm != "MD5" && f.algorithm != "MD5-SESS" {
		f.algorithm = "MD5"
	}

	if f.filePath == "" {
		argsError := make([]string, 0, 3)
		if f.sipSimpleUsername == "" {
			argsError = append(argsError, "SIP username is required!")
		}
		if f.sipSimplePassword == "" {
			argsError = append(argsError, "SIP password is required!")
		}
		if f.sipSimpleUri == "" {
			argsError = append(argsError, "SIP Server Address(uri) is required!")
		}
		if len(argsError) > 0 {
			note := []string{"\nRequired args:", strings.Join(argsError[:], "\n")}
			log.Fatal(strings.Join(note[:], "\n"))
		}
	} else {
		if !strings.HasPrefix(f.filePath, "/") {
			var pwd string
			pwd, _ = os.Getwd()
			f.filePath = fmt.Sprintf("%s/%s", pwd, f.filePath)
		}
	}
}

func main() {
	go logger(logChan)
	defer close(logChan)
	var flags flagType
	getFlags(&flags)

	if flags.filePath == "" {
		var c Connection = &connection{
			Username: flags.sipSimpleUsername,
			Password: flags.sipSimplePassword,
			Uri:      flags.sipSimpleUri,
			Port:     flags.sipSimplePort,
			Protocol: flags.protocol,
		}

		connectionList = append(connectionList, c)
	} else {
		connectionList = parseJsonFile(flags.filePath)
	}

	validateConnectionsCh := validateConnections(connectionList)

	for i := range validateConnectionsCh {
		// TODO: send to proxy function to send request in new go routines
		if i.GetObj().InviteDest != nil {
			fmt.Println(*i.GetObj().InviteDest)
		}
	}

	//for _, conn := range connectionList {
	//
	//}

	//	if protocol == "tcp" {
	//		err = socket.tcpDial()
	//		if err != nil {
	//			log.Fatal(fmt.Sprintf("Can not OPen client TCP socket in %s", clientIP))
	//		}
	//	} else {
	//		err = socket.udpDial()
	//		if err != nil {
	//			log.Fatal(fmt.Sprintf("Can not OPen client UDP socket in %s", clientIP))
	//		}
	//	}
	//
	//	if mode == "REGISTER" {
	//		logChan <- logMsg{
	//			level: 1,
	//			msg:   fmt.Sprintf("Starting send register packets with connection(%s)", protocol),
	//		}
	//		sendRegister(socket)
	//	}
	//}

}
