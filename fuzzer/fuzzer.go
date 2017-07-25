package fuzzer

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type StatusCodes []int

func (s StatusCodes) Has(x int) bool {
	for _, v := range s {
		if x == v {
			return true
		}
	}
	return false
}

type FuzzSettings struct {
	Tokens []string

	UrlAddress   string
	UrlTag       string
	BodyContains string

	FollowRedirect bool
	UserAgent      string
	Cookies        []http.Cookie

	ShowCodes StatusCodes
	HideCodes StatusCodes
}

type FuzzResponse struct {
	Token string

	Header     http.Header
	Status     string
	StatusCode int
	Location   string

	ContentLength int
	Body          string
	BodyContains  bool
}

type Fuzzer struct {
	FuzzSettings

	Throttler  chan int
	WaitGroup  *sync.WaitGroup
	HttpClient *http.Client
	Responses  chan FuzzResponse
}

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
	return &Fuzzer{}
}

func (f *Fuzzer) Start() {
	f.Responses = make(chan FuzzResponse, len(f.Tokens))
	for _, token := range f.Tokens {
		f.Push()
		go f.Check(token)
	}
}

func (f *Fuzzer) Check(token string) {
	url, tag := f.UrlAddress, f.UrlTag
	url = strings.Replace(url, tag, token, 1)

	response, err := f.request(url)
	if err != nil {
		f.Pop()
		return
	}
	defer response.Body.Close()

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

	var contains bool
	i := strings.Index(string(body), f.BodyContains)
	{
		if f.BodyContains != "" && i != -1 {
			contains = true
		}
	}

	contentLength := int(response.ContentLength)
	if response.ContentLength == -1 {
		contentLength = len(body)
	}

	res := FuzzResponse{
		Token: token,

		Header:     response.Header,
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Location:   response.Header.Get("Location"),

		ContentLength: contentLength,
		Body:          string(body),
		BodyContains:  contains,
	}

	f.Print(&res)
	f.Responses <- res

	f.Pop()
}

func (f *Fuzzer) request(url string) (resp *http.Response, err error) {
	if f.HttpClient == nil {
		f.HttpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	if !f.FollowRedirect {
		f.HttpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", f.UserAgent)

	return f.HttpClient.Do(req)
}
