package fuzzer

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func (f *Fuzzer) Push() {
	f.Throttler <- 1
	f.WaitGroup.Add(1)
}

func (f *Fuzzer) Pop() {
	f.Tokens = f.Tokens[1:]
	if len(f.Tokens) == 0 {
		close(f.Responses)
	}
	<-f.Throttler
	f.WaitGroup.Done()
}

func New() *Fuzzer {
	rand.Seed(time.Now().UTC().UnixNano())
	return &Fuzzer{
		HttpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (f *Fuzzer) Start() {
	f.Responses = make(chan FuzzResponse, len(f.Tokens))

	if !f.FollowRedirect {
		f.HttpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	for _, token := range f.Tokens {
		f.Push()
		go f.Check(token)

		sleepDuration := rand.Float64()*(f.Sleep.Max-f.Sleep.Min) + f.Sleep.Min
		time.Sleep(time.Duration(1000*sleepDuration) * time.Millisecond)
	}
}

func (f *Fuzzer) Check(token string) {
	fuzz := FuzzResponse{
		Token: token,
	}

	url, tag := f.UrlAddress, f.UrlTag
	url = strings.Replace(url, tag, token, -1)

	response, err := f.request(url, token)
	if err != nil {
		f.Pop()
		return
	}

	fuzz.Header = response.Header
	fuzz.Status, fuzz.StatusCode = response.Status, response.StatusCode
	fuzz.Location = response.Header.Get("Location")

	if len(f.HideCodes) != 0 && f.HideCodes.Has(response.StatusCode) {
		f.Pop()
		return
	}

	if len(f.ShowCodes) != 0 && !f.ShowCodes.Has(response.StatusCode) {
		f.Pop()
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		f.Pop()
		return
	}
	response.Body.Close()
	fuzz.Body = string(body)

	var contains bool
	if f.BodyContains != "" && strings.Index(string(body), f.BodyContains) != -1 {
		contains = true
	}
	fuzz.BodyContains = contains

	contentLength := int(response.ContentLength)
	if response.ContentLength == -1 {
		contentLength = len(body)
	}
	fuzz.ContentLength = contentLength

	f.Print(&fuzz)
	f.Responses <- fuzz

	f.Pop()
}

func (f *Fuzzer) request(urlAddr, token string) (resp *http.Response, err error) {
	var req *http.Request

	if f.Method == "POST" {
		req, err = http.NewRequest(f.Method, urlAddr, strings.NewReader(f.PostData.Encode()))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(f.Method, urlAddr, nil)
		if err != nil {
			return
		}
	}

	req.Header.Set("User-Agent", f.UserAgent)
	for k, v := range f.Headers {
		if v == f.UrlTag {
			v = token
		}
		req.Header.Set(k, v)
	}

	return f.HttpClient.Do(req)
}
