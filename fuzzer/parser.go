package fuzzer

import (
	"bufio"
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

func ParseHeadersFromString(x string) map[string]string {
	c := make(map[string]string)
	for _, v := range strings.Split(x, ",") {
		if s := strings.Split(v, "="); len(s) == 2 {
			c[s[0]] = s[1]
		}
	}
	return c
}

func ReadWordList(p string) (t []string, err error) {
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
