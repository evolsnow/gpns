package main

import (
	"fmt"
	"math/rand"
	"net/mail"
	"strings"
	"time"
)

//use mail's rfc2047 to encode any string
func encodeRFC2047(String string) string {
	address := mail.Address{Name: String, Address: ""}
	return strings.Trim(address.String(), "<@>")
}

//make mail message id
func makeMessageId(domain string) string {
	now := time.Now()
	utcDate := now.Format("20060102150405")
	rdm := rand.New(rand.NewSource(now.UnixNano()))
	randInt := rdm.Intn(100000)
	return fmt.Sprintf("<%s.%d@%s>", utcDate, randInt, domain)
}
