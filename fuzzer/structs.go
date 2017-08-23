package fuzzer

import (
	"net/http"
	"net/url"
	"sync"
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

type SleepInterval struct {
	Min float64
	Max float64
}

type FuzzSettings struct {
	Tokens []string

	UrlAddress   string
	UrlTag       string
	BodyContains string

	Method    string
	UserAgent string
	Headers   map[string]string
	PostData  url.Values
	//Cookies   []http.Cookie

	FollowRedirect bool
	ShowCodes      StatusCodes
	HideCodes      StatusCodes
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

	Throttler chan int
	WaitGroup *sync.WaitGroup

	Sleep SleepInterval

	HttpClient *http.Client
	Responses  chan FuzzResponse
}
