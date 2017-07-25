package fuzzer

import (
	"bufio"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

func ParseCodesFromString(x string) (c []int) {
	for _, v := range strings.Split(x, ",") {
		if n, err := strconv.Atoi(v); err == nil {
			c = append(c, n)
		}
	}
	sort.Ints(c)
	return
}

func ParseCookiesFromString(x string) (c []http.Cookie) {
	for _, v := range strings.Split(x, ",") {
		if s := strings.Split(v, "="); len(s) == 2 {
			c = append(c, http.Cookie{Name: s[0], Value: s[1]})
		}
	}
	return
}

func ReadWordlist(p string) (t []string, err error) {
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		t = append(t, s.Text())
	}
	return
}
