package main

import (
	"fmt"
	"strings"
)

func defaultInviteHeaderCreator(conn Connection) string {
	invite := fmt.Sprintf("INVITE sip:%s@%s:%d SIP/2.0\r\n",
		flags.inviteNumber, flags.uri, flags.port,
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
			"transport=%s>\r\nExpires: 0\r\n"+
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
			"Content-Length:   282", conn.Username, flags.uri,
	)

	return invite + via + from + to + useragent
}

func sendInvite(conn Connection) (Connection, error) {
	var header string
	for cSeq < 1 {
		header = defaultRegisterHeaderCreator(conn)
		err = conn.sendRequestToServer(header)

		if err == nil {
			break
		}
		cSeq += 1
		continue
	}
	return conn, nil
}
