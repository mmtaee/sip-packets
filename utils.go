package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	nonce      string
	realm      string
	cSeq       = 1
	cSeqText   = "REGISTER"
	callID     = uuid.New()
	branch     = createBranch()
	tag        = createTag()
	cnonce     = createCNonce()
	qop        string
	nonceCount = "00000001"
	route      string
)

func createBranch() string {
	prefix := "z9hG4bK"
	var branchCode string
	for i := 0; i < 16; i++ {
		branchCode += strconv.Itoa(rand.Intn(9))
	}
	return prefix + branchCode
}

func createTag() string {
	tagString := ""
	for i := 0; i < 8; i++ {
		tagString += strconv.Itoa(rand.Intn(9))
	}
	return tagString
}

func createCNonce() string {
	return "cnonce-" + uuid.New().String()
}

func makeHash(s string) string {
	hash := md5.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes)
	return hashStr
}

func nonceFinder(s string) {
	noncePattern := regexp.MustCompile(`nonce="([^"]+)"`)
	nonceMatch := noncePattern.FindStringSubmatch(s)
	if len(nonceMatch) > 1 {
		nonce = nonceMatch[1]
	}
}

func qopFinder(s string) {
	qopPattern := regexp.MustCompile(`qop="([^"]+)"`)
	qopMatch := qopPattern.FindStringSubmatch(s)
	if len(qopMatch) > 1 {
		qop = qopMatch[1]
		if strings.Contains(qop, "auth-int") {
			qop = "auth"
		}
	}
}

func realmFinder(s string) {
	realmPattern := regexp.MustCompile(`realm="([^"]+)"`)
	realmMatch := realmPattern.FindStringSubmatch(s)
	if len(realmMatch) > 1 {
		realm = realmMatch[1]
	}
}

func resultOutput(username string, result string) {
	_, err = os.Stat(flags.output)
	if errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(flags.output)
		if err != nil {
			logChan <- logMsg{
				level: 3,
				msg:   fmt.Sprintf("Error creating output file in path(%s): %s\n", flags.output, err),
			}
			return
		}
	}

	var file *os.File
	file, err = os.OpenFile(flags.output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	_, err = file.WriteString(fmt.Sprintf("%s | %s \n", username, result))
	if err != nil {
		logChan <- logMsg{
			level: 3,
			msg:   fmt.Sprintf("Error writing in file: %s", err),
		}
		return
	}
}

func cSeqTextFinder(s string) {
	cSeqTextPattern := regexp.MustCompile(`CSeq: \d+ (\w+)`)
	cSeqTextMatch := cSeqTextPattern.FindStringSubmatch(s)
	if len(cSeqTextMatch) > 1 {
		cSeqText = cSeqTextMatch[1]
	}
}

func routeFinder(s string) {
	routePattern := regexp.MustCompile(`Record-Route: (.*)`)
	routePatternMatch := routePattern.FindStringSubmatch(s)
	if len(routePatternMatch) > 1 {
		route = routePatternMatch[1]
	}
}
