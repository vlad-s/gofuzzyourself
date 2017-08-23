package flags

import (
	"log"
	"strings"
)

func CheckRequired(workers *int, urlAddr, fuzzFlag, wordList *string) {
	if *workers <= 0 {
		log.Fatalln("[ERROR] Workers must be a number bigger than zero")
	}

	switch "" {
	case *urlAddr:
		log.Fatalln("URL is not set, exiting.")
	case *fuzzFlag:
		log.Fatalln("Fuzz flag is not set, exiting.")
	case *wordList:
		log.Fatalln("Wordlist is not set, exiting.")
	}

	if strings.Index(*urlAddr, *fuzzFlag) == -1 {
		log.Fatalln("URL doesn't contain the fuzzing flag.")
	}
}

func CheckMethodAndData(httpMethod, postData *string) {
	*httpMethod = strings.ToUpper(*httpMethod)
	if *httpMethod != "GET" && *httpMethod != "HEAD" && *httpMethod != "POST" {
		log.Fatalln("[ERROR] Method should be either GET, HEAD, or POST.")
	}

	if *httpMethod == "POST" && *postData == "" {
		log.Println("[WARNING] Using POST method but no data to post!")
	}

	if *httpMethod != "POST" && *postData != "" {
		log.Printf("[WARNING] Setting post data but using %v method!", *httpMethod)
	}
}

func CheckFilters(showCodes, hideCodes *string) {
	if *showCodes != "" && *hideCodes != "" {
		log.Fatalln("[ERROR] `-show` and `-hide` flags are mutually exclusive.")
	}
}

func CheckSleep(sleep, sleepMin, sleepMax *float64) {
	if *sleep > 0 && (*sleepMin > 0 || *sleepMax > 0) {
		log.Fatalln("`-sleep` and `-sleep-min`/`-sleep-max` flags are mutually exclusive.")
	}

	if *sleepMin > *sleepMax {
		log.Fatalln("`-sleep-min` can't be higher than `-sleep-max`.")
	}

	if *sleep < 0 || *sleepMin < 0 || *sleepMax < 0 {
		log.Fatalln("Can't sleep for a negative period.")
	}

	if *sleep > 0 {
		*sleepMin = *sleep
		*sleepMax = *sleep
	}
}
