package main

import (
	"errors"
	"fmt"
	"strings"
)

func makeRegisterPacketHash(obj Connection) string {
	a1 := fmt.Sprintf("%s:%s:%s", obj.Username, realm, obj.Password)
	ha1 := makeHash(a1)
	if flags.algorithm == "MD5-SESS" {
		a1 = fmt.Sprintf("%s:%s:%s", ha1, nonce, cnonce)
		ha1 = makeHash(a1)
	}
	a2 := "REGISTER:sip:" + flags.uri
	ha2 := makeHash(a2)
	if qop == "auth-int" {
		// qop type of authentication : auth-int or int
		hashEntityBody := "d41d8cd98f00b204e9800998ecf8427e"
		a2 = fmt.Sprintf("REGISTER:%s:%s", flags.uri, hashEntityBody)
		ha2 = makeHash(a2)
	}
	var res string
	if qop != "" {
		res = fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			ha1, nonce, nonceCount, cnonce, qop, ha2,
		)
	} else {
		res = fmt.Sprintf("%s:%s:%s",
			ha1, nonce, ha2,
		)
	}
	return makeHash(res)
}

func defaultRegisterHeaderCreator(conn Connection) string {
	register := fmt.Sprintf("REGISTER sip:%s:%d;transport=%s SIP/2.0\r\n", flags.uri, flags.port, flags.protocol)
	via := fmt.Sprintf(
		"Via: SIP/2.0/%s %s:%d;branch=%s;rport\r\nMax-Forwards: 70\r\n",
		strings.ToUpper(flags.protocol), clientIP, conn.ClientPort, branch,
	)
	from := fmt.Sprintf(
		"From: \"%s\" <sip:%s@%s>;tag=%s\r\n",
		conn.Username, conn.Username, flags.uri, tag,
	)
	to := fmt.Sprintf(
		"To: <sip:%s@%s>\r\nContact: <sip:%s@%s:%d;"+
			"transport=%s>\r\nExpires: 30\r\n"+
			"Call-ID: %s\r\nAccept: application/sdp\r\nContent-Length: 0\r\n",
		conn.Username, flags.uri, conn.Username, clientIP, conn.ClientPort, flags.protocol, callID,
	)
	return register + via + from + to
}

func nonceHeaderCreator(conn Connection, defaultHeader string) string {
	var authorization string
	responseHash := makeRegisterPacketHash(conn)
	uri := flags.uri
	if flags.strategy == "INVITE" {
		uri = fmt.Sprintf("%s@%s", flags.inviteNumber, flags.uri)
	}
	defaultHeader += fmt.Sprintf("CSeq: %d %s\r\n", cSeq, flags.strategy)
	if qop != "" {
		authorization = fmt.Sprintf("Authorization:  Digest realm=\"%s\", nonce=\"%s\","+
			" username=\"%s\", uri=\"sip:%s\", response=\"%s\", nc=%s, algorithm=%s",
			realm, nonce, conn.Username, uri, responseHash, nonceCount, flags.algorithm,
		)
		if cnonce != "" {
			authorization += fmt.Sprintf(", cnonce=\"%s\"", cnonce)
		}
		authorization += ", qop=" + qop
	} else {
		authorization = fmt.Sprintf(
			"Authorization:  Digest realm=\"%s\", username=\"%s\", "+
				"response=\"%s\", nonce=\"%s\", uri=\"sip:%s\", algorithm=%s",
			realm, conn.Username, responseHash, nonce, uri, flags.algorithm,
		)
	}

	defaultHeader += authorization
	return defaultHeader
}

func sendRegister(conn Connection) (Connection, error) {
	var header string
	cSeq = 0
	for cSeq < 5 {
		header = defaultRegisterHeaderCreator(conn)
		if nonce != "" {
			header = nonceHeaderCreator(conn, header)
		} else {
			header += fmt.Sprintf("CSeq: %d %s\r\n", cSeq, cSeqText)
		}
		err = conn.sendRequestToServer(header)
		if err != nil || conn.Status == 2 {
			break
		}
		cSeq += 1
	}
	var finalMsg logMsg
	if code, _ := conn.getResult(); code == 2 {
		finalMsg = logMsg{
			level: 1,
			msg:   fmt.Sprintf("User(%s) registered successfully on sip server(%s)", conn.Username, flags.uri),
		}
	} else {
		conn.Status = 23
		finalMsg = logMsg{
			level: 1,
			msg: fmt.Sprintf("User(%s) registered failed on sip server(%s) with result: %s (status: %d)",
				conn.Username, flags.uri,
				conn.Status.String(), conn.Status,
			),
		}
	}
	if flags.verbose {
		logChan <- finalMsg
	} else {
		forceLogger(finalMsg)
	}
	if conn.Status == 2 {
		return conn, nil
	}
	return conn, errors.New("registered failed")
}
