package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

var (
	nonce      string
	realm      string
	cSeq       = 0
	callID     = uuid.New()
	branch     = createBranch()
	tag        = createTag()
	cnonce     = createCNonce()
	qop        string
	nonceCount = "00000001"
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
