package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"fmt"
	"os/signal"

	"github.com/vlad-s/gofuzzyourself/fuzzer"
)

var (
	help    = flag.Bool("h", false, "Show the help message")
	verbose = flag.Bool("v", false, "Verbose output")
	workers = flag.Int("workers", 32, "How many spawned workers")

	urlFlag      = flag.String("U", "", "The `URL` to fuzz")
	fuzzFlag     = flag.String("F", "$fuzz$", "The `flag` to use in fuzzing")
	wordlistFlag = flag.String("W", "", "The `wordlist` to use in fuzzing")
	//cookiesFlag  = flag.String("C", "", "`Cookies` to use, separated by a semicolon")
	//postFlag     = flag.String("D", "", "The post `data` to use in fuzzing")
	//headerFlag   = flag.String("H", "", "The `header` to use in fuzzing")

	containsFlag  = flag.String("contains", "", "Search the body for the specified `string`")
	followFlag    = flag.Bool("follow", false, "Follow or not redirects")
	userAgentFlag = flag.String("user-agent", "-", "The `User-Agent` to use")

	showFlag = flag.String("show", "", "Show only the status `codes`, separated by comma")
	hideFlag = flag.String("hide", "", "Hide the status `codes`, separated by comma")
)

var (
	Throttler chan int

	FuzzTokens []string
	ShowCodes  fuzzer.StatusCodes
	HideCodes  fuzzer.StatusCodes
	Cookies    []http.Cookie
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
	Throttler = make(chan int, *workers)
	if *verbose {
		log.Printf("[DEBUG] Setting the throttler as %T(%d, %d)", Throttler, len(Throttler), cap(Throttler))
	}

	switch "" {
	case *urlFlag:
		log.Fatalln("[ERROR] URL is not set, exiting.")
	case *fuzzFlag:
		log.Fatalln("[ERROR] Fuzz flag is not set, exiting.")
	case *wordlistFlag:
		log.Fatalln("[ERROR] Wordlist is not set, exiting.")
	}

	var err error
	FuzzTokens, err = fuzzer.ReadWordlist(*wordlistFlag)
	if err != nil {
		log.Fatalln("[ERROR] Can't read wordlist, exiting.")
	}

	if strings.Index(*urlFlag, *fuzzFlag) == -1 {
		log.Fatalln("[ERROR] URL doesn't contain the fuzzing flag, exiting.")
	}

	if *showFlag != "" && *hideFlag != "" {
		log.Fatalln("[ERROR] `-show` and `-hide` flags are mutually exclusive, exiting.")
	}

	ShowCodes = fuzzer.ParseCodesFromString(*showFlag)
	if *verbose {
		log.Println("[DEBUG] Show only status codes:", ShowCodes)
	}

	HideCodes = fuzzer.ParseCodesFromString(*hideFlag)
	if *verbose {
		log.Println("[DEBUG] Hide status codes:", HideCodes)
	}

	/*Cookies = fuzzer.ParseCookiesFromString(*cookiesFlag)
	if *verbose {
		log.Println("[DEBUG] Use cookies:", Cookies)
	}*/
}

func main() {
	parseFlags()

	f := fuzzer.New()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("[KILL] We got interrupt")
			fmt.Printf("We only got %d responses until now.\n", len(f.Responses))
			os.Exit(0)
		}
	}()

	var wg sync.WaitGroup
	f.WaitGroup = &wg
	f.Throttler = Throttler

	f.FuzzSettings = fuzzer.FuzzSettings{
		Tokens: FuzzTokens,

		UrlAddress:   *urlFlag,
		UrlTag:       *fuzzFlag,
		BodyContains: *containsFlag,

		FollowRedirect: *followFlag,
		UserAgent:      *userAgentFlag,
		Cookies:        Cookies,

		ShowCodes: ShowCodes,
		HideCodes: HideCodes,
	}

	if *verbose {
		log.Printf("[DEBUG] Fuzzer: %+v", f)
	}

	f.PrintHeader()
	f.Start()

	wg.Wait()
}
