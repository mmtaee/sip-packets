package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

var (
	protocol     string
	err          error
	sipUri       string
	sipPort      int
	sipUsername  string
	sipPassword  string
	algorithm    string
	mode         string
	verbose      bool
	vVerbose     bool
	isRegistered            = false
	logChan                 = make(chan logMsg)
	socket       Connection = &connection{}
)

func getFlags() {
	argsError := make([]string, 0, 3)
	flag.StringVar(&sipUsername, "username", "", "SIP username")
	flag.StringVar(&sipPassword, "password", "", "SIP Password")
	flag.StringVar(&sipUri, "uri", "", "SIP Server Address(Destination Address)")
	flag.IntVar(&sipPort, "port", 5060, "SIP Server Port")
	flag.StringVar(&protocol, "protocol", "tcp", "SIP Server Auth Protocol")
	flag.StringVar(&algorithm, "algorithm", "md5", "SIP Server Auth Algorithm(MD5 or MD5-SESS)")
	flag.StringVar(&mode, "mode", "register", "SIP Packet Type(register=1, invite=2)")
	flag.BoolVar(&vVerbose, "v", false, "Verbose Mode")
	flag.BoolVar(&verbose, "verbose", false, "Verbose Mode")
	flag.Parse()
	protocol = strings.ToLower(protocol)
	if protocol != "tcp" && protocol != "udp" {
		protocol = "tcp"
		logChan <- logMsg{
			level: 1,
			msg:   "protocol not valid. set to default(tcp)",
		}
	}
	if verbose || vVerbose {
		verbose = true
	}
	if algorithm != "MD5" && algorithm != "MD5-SESS" {
		algorithm = "MD5"
	}
	if sipUsername == "" {
		argsError = append(argsError, "SIP username is required!")
	}
	if sipPassword == "" {
		argsError = append(argsError, "SIP password is required!")
	}
	if sipUri == "" {
		argsError = append(argsError, "SIP Server Address(uri) is required!")
	}
	if len(argsError) > 0 {
		note := []string{"\nRequired args:", strings.Join(argsError[:], "\n")}
		log.Fatal(strings.Join(note[:], "\n"))
	}
	modes := []string{"REGISTER", "INVITE"}
	mode = strings.ToUpper(mode)
	if !strings.Contains(strings.Join(modes, ","), mode) {
		mode = "REGISTER"
		logChan <- logMsg{
			level: 1,
			msg:   "mode not set. set to default(REGISTER)",
		}
	}
}

func main() {
	getFlags()

	go logger(logChan)
	defer close(logChan)

	if protocol == "tcp" {
		err = socket.tcpDial()
		if err != nil {
			log.Fatal(fmt.Sprintf("Can not OPen client TCP socket in %s", clientIP))
		}
	} else {
		err = socket.udpDial()
		if err != nil {
			log.Fatal(fmt.Sprintf("Can not OPen client UDP socket in %s", clientIP))
		}
	}

	if mode == "REGISTER" {
		logChan <- logMsg{
			level: 1,
			msg:   fmt.Sprintf("Starting send register packets with connection(%s)", protocol),
		}
		sendRegister(socket)
	}

}
