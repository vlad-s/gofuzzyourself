package fuzzer

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	printFormat string
)

func makeFormat(h []string) (s string, i []interface{}) {
	for _, v := range h {
		s += fmt.Sprintf("| %%%dv ", len(v))
		i = append(i, v)
	}
	s = strings.Trim(s, " ") + "\n"
	return
}

func limitToken(t string, s int) string {
	if len(t) > s {
		maxLen := (s - 3) / 2
		t = t[:maxLen] + "..." + t[len(t)-maxLen:]
	}
	return t
}

func (f *Fuzzer) PrintHeader() {
	log.Printf("[INFO] Preparing to check %d URLs, stand by...", len(f.Tokens))

	headers := []string{"Status", "Length"}
	if f.BodyContains != "" {
		headers = append(headers, "Contains")
	}
	if f.FollowRedirect == false {
		headers = append(headers, "Location")
	}
	headers = append(headers, "Token")

	data := []interface{}{}
	printFormat, data = makeFormat(headers)
	header := fmt.Sprintf(printFormat, data...)

	fmt.Fprintf(os.Stdout, header)
	for i := 3; i < len(header)+2; i++ {
		fmt.Printf("-")
	}
	fmt.Printf("\n")
}

func (f *Fuzzer) Print(r *FuzzResponse) {
	r.Token = limitToken(r.Token, 60)
	r.Location = limitToken(r.Location, 60)

	body := []interface{}{r.StatusCode, r.ContentLength}
	if f.BodyContains != "" {
		body = append(body, r.BodyContains)
	}
	if f.FollowRedirect == false {
		body = append(body, r.Location)
	}
	body = append(body, r.Token)
	fmt.Fprintf(os.Stdout, printFormat, body...)
}
