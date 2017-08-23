package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/vlad-s/gofuzzyourself/flags"
	"github.com/vlad-s/gofuzzyourself/fuzzer"
)

var (
	workers = flag.Int("workers", 32, "How many spawned workers")

	fuzzUrl       = flag.String("url", "", "The `URL` to fuzz")
	fuzzFlag      = flag.String("flag", "$fuzz$", "The `flag` to use")
	wordList      = flag.String("wordlist", "", "The `wordlist` to use")
	httpMethod    = flag.String("method", "GET", "The HTTP `method` to use")
	httpHeaders   = flag.String("headers", "", "The `headers` to use, separated by comma")
	httpUserAgent = flag.String("user-agent", "-", "The `User-Agent` to use")
	postData      = flag.String("data", "", "The `post data` to use, fields separated by comma")
	//cookiesFlag  = flag.String("C", "", "`Cookies` to use, separated by a semicolon")

	contains = flag.String("contains", "", "Search the body for the specified `string`")
	follow   = flag.Bool("follow", false, "Follow or not redirects")

	showCodes = flag.String("show", "", "Show only the status `codes`, separated by comma")
	hideCodes = flag.String("hide", "", "Hide the status `codes`, separated by comma")

	sleep    = flag.Float64("sleep", 0, "How many seconds to sleep between requests (default 0)\n\tEqual to -sleep-min X -sleep-max X")
	sleepMin = flag.Float64("sleep-min", 0, "Inferior limit of the time range (default 0)")
	sleepMax = flag.Float64("sleep-max", 0, "Superior limit of the time range (default 0)")
)

var (
	err error
	wg  sync.WaitGroup

	Fuzzer     *fuzzer.Fuzzer
	FuzzTokens []string
)

func parseFlags() {
	flag.Parse()

	flags.CheckRequired(workers, fuzzUrl, fuzzFlag, wordList)
	flags.CheckMethodAndData(httpMethod, postData)
	flags.CheckFilters(showCodes, hideCodes)
	flags.CheckSleep(sleep, sleepMin, sleepMax)

	FuzzTokens, err = fuzzer.ReadWordList(*wordList)
	if err != nil {
		log.Fatalln("[ERROR] Can't read wordlist.")
	}
}

func init() {
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	parseFlags()
	Fuzzer = fuzzer.New()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("[KILL] We got ^C")
			log.Printf("We only got %d responses until now.\n", len(Fuzzer.Responses))
			os.Exit(0)
		}
	}()
}

func main() {
	Fuzzer.WaitGroup = &wg
	Fuzzer.Throttler = make(chan int, *workers)

	Fuzzer.Sleep = fuzzer.SleepInterval{
		Min: *sleepMin,
		Max: *sleepMax,
	}

	Fuzzer.FuzzSettings = fuzzer.FuzzSettings{
		Tokens: FuzzTokens,

		UrlAddress:   *fuzzUrl,
		UrlTag:       *fuzzFlag,
		BodyContains: *contains,

		FollowRedirect: *follow,
		Headers:        fuzzer.ParseHeaders(*httpHeaders),
		Method:         *httpMethod,
		PostData:       fuzzer.ParsePostData(*postData),
		UserAgent:      *httpUserAgent,

		ShowCodes: fuzzer.ParseStatusCodes(*showCodes),
		HideCodes: fuzzer.ParseStatusCodes(*hideCodes),
	}

	Fuzzer.PrintHeader()
	Fuzzer.Start()

	wg.Wait()
}
