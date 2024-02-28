package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func defaultInviteHeaderCreator(conn Connection, sipTitle string) string {
	invite := fmt.Sprintf("%s sip:%s@%s:%d SIP/2.0\r\n",
		sipTitle, flags.inviteNumber, flags.uri, flags.port,
	)

	via := fmt.Sprintf(
		"Via: SIP/2.0/%s %s:%d;branch=%s;rport;alias\r\nMax-Forwards: 70\r\n",
		strings.ToUpper(flags.protocol), clientIP, conn.ClientPort, branch,
	)

	from := fmt.Sprintf(
		"From: \"%s\" <sip:%s@%s>;tag=%s\r\n",
		conn.Username, conn.Username, flags.uri, tag,
	)

	to := fmt.Sprintf(
		"To: <sip:%s@%s>\r\nContact: <sip:%s@%s:%d;"+
			"transport=%s>\r\n"+ // Expires: 0
			"Call-ID: %s\r\nAccept: application/sdp\r\nContent-Length: 0\r\n",
		flags.inviteNumber, flags.uri, conn.Username, clientIP, conn.ClientPort, flags.protocol, callID,
	)

	useragent := fmt.Sprintf(
		"User-Agent: Grandstream Wave 1.2.14\r\n"+
			"Privacy: none\r\n"+
			"P-Preferred-Identity: <sip:%s@%s>\r\n"+
			"Supported: replaces, path, timer, eventlist\r\n"+
			"Allow: INVITE, ACK, OPTIONS, CANCEL, BYE, SUBSCRIBE, NOTIFY, INFO, REFER, UPDATE, MESSAGE\r\n"+
			"Content-Type: application/sdp\r\n"+
			"Accept: application/sdp, application/dtmf-relay\r\n"+
			"Content-Length:   282\r\n", conn.Username, flags.uri,
	)

	return invite + via + from + to + useragent
}

func sendInvite(conn Connection) (Connection, error) {
	var header string
	var ringing bool
	var finalMsg logMsg
	// TODO: move RINGING to flags
	if os.Getenv("RINGING") != "" {
		ringing = false
	} else {
		ringing = true
	}
	for cSeq < 5 {

		fmt.Println(conn.Status)
		fmt.Println("#################33\n\n")

		header = defaultInviteHeaderCreator(conn, "INVITE")
		if conn.Status == 3 { // 401 Unauthorized
			header = nonceHeaderCreator(conn, header)
			fmt.Println(header)
		}
		if conn.Status == 8 { // 183 Session Progress
			if ringing {
				// scenario 1: send ack and then bye
				header = defaultInviteHeaderCreator(conn, "CANCEL")
				header += fmt.Sprintf("Route: %s", route)
			} else {
				// scenario 2: send cancel only
				cSeqText = "CANCEL"
				header = defaultInviteHeaderCreator(conn, "CANCEL")
			}
		}
		if conn.Status == 2 { // 200 ok
			// BYE response
		}

		if conn.Status == 9 { // 200 Canceled

			finalMsg = logMsg{
				level: 1,
				msg:   fmt.Sprintf("User(%s) registered successfully on sip server(%s)", conn.Username, flags.uri),
			}
			break

		}
		header += fmt.Sprintf("CSeq: %d %s\r\n", cSeq, cSeqText)
		err = conn.sendRequestToServer(header)
		if err != nil {
			break
		} else {
			log.Println(err)
		}
		cSeq += 1
		continue
	}
	if flags.verbose {
		logChan <- finalMsg
	} else {
		forceLogger(finalMsg)
	}
	return conn, nil
}
