package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"os/signal"

	"github.com/vlad-s/gofuzzyourself/fuzzer"
)

var (
	help    = flag.Bool("h", false, "Show the help message")
	verbose = flag.Bool("v", false, "Verbose output")
	workers = flag.Int("workers", 32, "How many spawned workers")

	urlFlag  = flag.String("U", "", "The `URL` to fuzz")
	fuzzFlag = flag.String("F", "$fuzz$", "The `flag` to use in fuzzing")
	wordList = flag.String("W", "", "The `wordlist` to use in fuzzing")
	headers  = flag.String("H", "", "The `headers` to use in fuzzing, separated by comma")
	//cookiesFlag  = flag.String("C", "", "`Cookies` to use, separated by a semicolon")
	//postFlag     = flag.String("D", "", "The post `data` to use in fuzzing")

	contains  = flag.String("contains", "", "Search the body for the specified `string`")
	follow    = flag.Bool("follow", false, "Follow or not redirects")
	userAgent = flag.String("user-agent", "-", "The `User-Agent` to use")

	showCodes = flag.String("show", "", "Show only the status `codes`, separated by comma")
	hideCodes = flag.String("hide", "", "Hide the status `codes`, separated by comma")
)

var (
	wg sync.WaitGroup

	FuzzTokens []string
)

func parseFlags() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *workers <= 0 {
		log.Fatalln("[ERROR] Workers must be a number bigger than zero")
	}

	switch "" {
	case *urlFlag:
		log.Fatalln("[ERROR] URL is not set, exiting.")
	case *fuzzFlag:
		log.Fatalln("[ERROR] Fuzz flag is not set, exiting.")
	case *wordList:
		log.Fatalln("[ERROR] Wordlist is not set, exiting.")
	}

	var err error
	FuzzTokens, err = fuzzer.ReadWordList(*wordList)
	if err != nil {
		log.Fatalln("[ERROR] Can't read wordlist, exiting.")
	}

	if strings.Index(*urlFlag, *fuzzFlag) == -1 {
		log.Fatalln("[ERROR] URL doesn't contain the fuzzing flag, exiting.")
	}

	if *showCodes != "" && *hideCodes != "" {
		log.Fatalln("[ERROR] `-show` and `-hide` flags are mutually exclusive, exiting.")
	}
}

func main() {
	parseFlags()

	f := fuzzer.New()
	notifyInterrupt(f)

	f.WaitGroup = &wg
	f.Throttler = make(chan int, *workers)

	f.FuzzSettings = fuzzer.FuzzSettings{
		Tokens: FuzzTokens,

		UrlAddress:   *urlFlag,
		UrlTag:       *fuzzFlag,
		BodyContains: *contains,

		FollowRedirect: *follow,
		Headers:        fuzzer.ParseHeadersFromString(*headers),
		UserAgent:      *userAgent,

		ShowCodes: fuzzer.ParseCodesFromString(*showCodes),
		HideCodes: fuzzer.ParseCodesFromString(*hideCodes),
	}

	if *verbose {
		log.Printf("[DEBUG] Fuzzer: %+v", f)
	}

	f.PrintHeader()
	f.Start()

	wg.Wait()
}

func notifyInterrupt(f *fuzzer.Fuzzer) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("[KILL] We got ^C")
			log.Printf("We only got %d responses until now.\n", len(f.Responses))
			os.Exit(0)
		}
	}()
}
