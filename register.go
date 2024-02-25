package main

//
//import (
//	"fmt"
//	"strings"
//)
//
//func makeRegisterPacketHash() string {
//	a1 := fmt.Sprintf("%s:%s:%s", sipUsername, realm, sipPassword)
//	ha1 := makeHash(a1)
//	if algorithm == "MD5-SESS" {
//		a1 = fmt.Sprintf("%s:%s:%s", ha1, nonce, cnonce)
//		ha1 = makeHash(a1)
//	}
//	a2 := "REGISTER:sip:" + sipUri
//	ha2 := makeHash(a2)
//	if qop == "auth-int" {
//		// qop type of authentication : auth-int or int
//		hashEntityBody := "d41d8cd98f00b204e9800998ecf8427e"
//		a2 = fmt.Sprintf("REGISTER:%s:%s", sipUri, hashEntityBody)
//		ha2 = makeHash(a2)
//	}
//	var res string
//	if qop != "" {
//		res = fmt.Sprintf("%s:%s:%s:%s:%s:%s",
//			ha1, nonce, nonceCount, cnonce, qop, ha2,
//		)
//	} else {
//		res = fmt.Sprintf("%s:%s:%s",
//			ha1, nonce, ha2,
//		)
//	}
//	return makeHash(res)
//}
//
//func defaultRegisterHeaderCreator() string {
//	register := fmt.Sprintf("REGISTER sip:%s:%d;transport=%s SIP/2.0\r\n", sipUri, sipPort, protocol)
//	via := fmt.Sprintf(
//		"Via: SIP/2.0/%s %s:%d;branch=%s;rport\r\nMax-Forwards: 70\r\n",
//		strings.ToUpper(protocol), clientIP, clientPort, branch,
//	)
//	from := fmt.Sprintf(
//		"From: \"%s\" <sip:%s@%s>;tag=%s\r\n",
//		sipUsername, sipUsername, sipUri, tag,
//	)
//	to := fmt.Sprintf(
//		"To: <sip:%s@%s>\r\nContact: <sip:%s@%s:%d;"+
//			"transport=%s>\r\nExpires: 0\r\n"+
//			"Call-ID: %s\r\nAccept: application/sdp\r\nContent-Length: 0\r\n",
//		sipUsername, sipUri, sipUsername, clientIP, clientPort, protocol, callID,
//	)
//	return register + via + from + to
//}
//
//func nonceHeaderCreator(defaultHeader string) string {
//	var authorization string
//	responseHash := makeRegisterPacketHash()
//	defaultHeader += fmt.Sprintf("CSeq: %d REGISTER\r\n", cSeq)
//	if qop != "" {
//		authorization = fmt.Sprintf("Authorization:  Digest realm=\"%s\", nonce=\"%s\","+
//			" username=\"%s\", uri=\"sip:%s\", response=\"%s\", nc=%s",
//			realm, nonce, sipUsername, sipUri, responseHash, nonceCount,
//		)
//		if cnonce != "" {
//			authorization += fmt.Sprintf(", cnonce=\"%s\"", cnonce)
//		}
//		authorization += ", qop=" + qop
//	} else {
//		authorization = fmt.Sprintf(
//			"Authorization:  Digest realm=\"%s\", username=\"%s\", "+
//				"response=\"%s\", nonce=\"%s\", uri=\"sip:%s\"",
//			realm, sipUsername, responseHash, nonce, sipUri,
//		)
//	}
//	defaultHeader += authorization
//	return defaultHeader
//}
//
//func sendRegister(conn Connection) {
//	var header string
//	for cSeq < 5 {
//		header = defaultRegisterHeaderCreator(conn)
//		if nonce != "" {
//			header = nonceHeaderCreator(header)
//		} else {
//			header += fmt.Sprintf("CSeq: %d REGISTER\r\n", cSeq)
//		}
//		header += "\r\n\r\n"
//		logChan <- logMsg{
//			level: 1,
//			msg:   "Sent Packet Header: \n\t" + strings.Replace(header, "\r\n", "\n\t", -1),
//		}
//		_, err = conn.Write([]byte(header))
//		if err != nil {
//			return
//		}
//
//		buffer := make([]byte, 1024)
//		var response int
//		response, err = conn.Read(buffer)
//		if err != nil {
//			logChan <- logMsg{
//				level: 2,
//				msg:   fmt.Sprintf("Error reading from response: %s", err),
//			}
//			return
//		}
//
//		logChan <- logMsg{
//			level: 1,
//			msg:   "Response from server: \n\t" + strings.Replace(string(buffer[:response]), "\r\n", "\n\t", -1),
//		}
//
//		sipResponseString := string(buffer[:response])
//
//		if strings.Contains(string(buffer[:response]), "200 OK") {
//			isRegistered = true
//			break
//		}
//
//		if strings.Contains(string(buffer[:response]), "100 Trying") {
//			logChan <- logMsg{
//				level: 1,
//				msg:   "100 Trying",
//			}
//		}
//
//		if strings.Contains(string(buffer[:response]), "401 Unauthorized") {
//			logChan <- logMsg{
//				level: 1,
//				msg:   "401 Unauthorized",
//			}
//			nonceFinder(sipResponseString)
//			qopFinder(sipResponseString)
//			realmFinder(sipResponseString)
//		}
//		cSeq += 1
//		logChan <- logMsg{
//			level: 1,
//			msg:   fmt.Sprintf("retrying cSeq(%d) ...", cSeq),
//		}
//		continue
//	}
//	var finalMsg logMsg
//	if isRegistered {
//		finalMsg = logMsg{
//			level: 1,
//			msg:   fmt.Sprintf("User(%s) registered successfully on sip server(%s)", sipUsername, sipUri),
//		}
//
//	} else {
//		finalMsg = logMsg{
//			level: 1,
//			msg:   fmt.Sprintf("User(%s) registered failed on sip server(%s)", sipUsername, sipUri),
//		}
//	}
//	if verbose {
//		logChan <- finalMsg
//	} else {
//		forceLogger(finalMsg)
//	}
//}
