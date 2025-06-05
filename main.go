package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	oast       = flag.String("oa", "", "Embed OAST link (e.g. .example.oast.online)")
	canary     = flag.String("c", "timestamp", "Canary mode: timestamp, custom string or range (1-10)")
	output     = flag.String("o", "requests.dreq", "Output file")
	suffix     = flag.String("s", "", "Suffix to add after token")
	prefix     = flag.String("p", "", "Prefix to add before token")
	delimiter  = flag.String("d", "", "Delimiter between token and OAST (e.g. timestamp.oast.me)")
	keepOrig   = flag.Bool("k", false, "Keep original value and append token")
	replaceAll = flag.Bool("a", false, "Replace all matching params")
	target     = flag.String("t", "", "Comma-separated list of parameters to target")
	inputFile  = flag.String("f", "", "Input file (default: stdin)")
	rawMode    = flag.Bool("r", false, "Parse raw HTTP requests")
	verbose    = flag.Bool("v", false, "Verbose output")
	quiet      = flag.Bool("q", false, "Quiet mode")
	logFile    = flag.String("l", "", "Log replaced lines to file")
	help       = flag.Bool("h", false, "Show help message")
)

var targetParams map[string]bool
var rangeCanary []string

func showBanner() {
	banner := `
   ______                            
  / ____/___ _____  ____ ________  __
 / /   / __ ` + "`" + `/ __ \/ __ ` + "`" + `/ ___/ / / /
/ /___/ /_/ / / / / /_/ / /  / /_/ / 
\____/\__,_/_/ /_/\__,_/_/   \__, /  
                            /____/   
	                 @rootplinix

`
	fmt.Println(banner)
}

func showHelp() {
	banner := `
Canary - Canary Token Injector for Bug Bounty Hunting
`
	fmt.Println(banner)
	helpText := `
Usage:
  canary [options] < input.txt

Options:
  -oa, --oast <OAST_URL>       Embed OAST link (e.g. .example.oast.online)
  -c,  --canary <MODE>         Use timestamp, custom string, or range (e.g. 1-100)
  -o,  --output <FILE>         Output file (default: requests.dreq)
  -s,  --suffix <SUFFIX>       Suffix to add after token (e.g. .tracker)
  -p,  --prefix <PREFIX>       Prefix to add before token
  -d,  --delimiter <CHAR>      Delimiter to use between token and OAST (e.g. timestamp.oast)
  -k,  --keep                  Keep original parameter value and append token
  -a,  --all                   Replace all matching parameters, not just the first
  -t,  --target <param,param>  Comma-separated list of parameters to target
  -f,  --file <FILE>           Read input from file instead of stdin
  -r,  --raw                   Parse raw HTTP requests
  -l,  --log <FILE>            Log replaced lines to file
  -v,  --verbose               Enable verbose output
  -q,  --quiet                 Suppress final summary output
  -h,  --help                  Show this help message

Examples:
  cat urls.txt | canary -oa .oast.me -c timestamp -o output.txt
  canary -f requests.txt -c mytoken -a -k -l changes.log
`
	fmt.Println(helpText)
	os.Exit(0)
}

func parseRange(rng string) []string {
	parts := strings.Split(rng, "-")
	if len(parts) != 2 {
		return []string{rng}
	}
	start, err1 := strconv.Atoi(parts[0])
	end, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || start > end {
		return []string{rng}
	}
	var out []string
	for i := start; i <= end; i++ {
		out = append(out, fmt.Sprintf("%d", i))
	}
	return out
}

func injectToken(line string, lineNum int, canaryIdx *int, lock *sync.Mutex) (string, string) {
	orig := line
	var path string
	if *rawMode {
		re := regexp.MustCompile(`(?i)^[A-Z]+ (/[^ ]*) HTTP`)
		matches := re.FindStringSubmatch(line)
		if len(matches) < 2 {
			return "", ""
		}
		path = matches[1]
	} else {
		path = line
	}

	if !strings.Contains(path, "?") {
		return "", ""
	}

	parts := strings.SplitN(path, "?", 2)
	base, paramStr := parts[0], parts[1]

	uParams, _ := url.ParseQuery(paramStr)
	token := ""

	lock.Lock()
	switch {
	case len(rangeCanary) > 0:
		token = rangeCanary[*canaryIdx%len(rangeCanary)]
		*canaryIdx++
	case *canary == "timestamp":
		token = fmt.Sprintf("%d%d", time.Now().Unix(), lineNum)
	default:
		token = *canary
	}
	lock.Unlock()

	if *oast != "" {
		if *delimiter != "" {
			token = fmt.Sprintf("https://%s%s%s", token, *delimiter, *oast)
		} else {
			token = fmt.Sprintf("https://%s%s", token, *oast)
		}
	}
	token = *prefix + token + *suffix

	changed := false
	for key, vals := range uParams {
		if len(targetParams) > 0 && !targetParams[key] {
			continue
		}
		for i := range vals {
			if *keepOrig {
				uParams[key][i] = uParams[key][i] + token
			} else {
				uParams[key][i] = token
			}
			changed = true
			if !*replaceAll {
				break
			}
		}
		if changed && !*replaceAll {
			break
		}
	}

	if !changed {
		return "", ""
	}

	encoded := uParams.Encode()
	final := fmt.Sprintf("%s?%s", base, encoded)
	return final, fmt.Sprintf("%s -> %s", orig, final)
}

func main() {
	flag.Parse()

	if *help {
		showHelp()
	}

	showBanner()

	// Check if no input supplied: no file flag and no stdin data
	fi, err := os.Stdin.Stat()
	if *inputFile == "" && (err != nil || (fi.Mode()&os.ModeCharDevice) != 0) && len(flag.Args()) == 0 {
		// No input file, and stdin is from terminal (no pipe), and no args
		fmt.Fprintln(os.Stderr, "[!] No input supplied. Please provide input via pipe or -f flag.")
		os.Exit(0)
}

	if *target != "" {
		targetParams = make(map[string]bool)
		for _, param := range strings.Split(*target, ",") {
			targetParams[param] = true
		}
	}
	if strings.Contains(*canary, "-") {
		rangeCanary = parseRange(*canary)
	}

	var input *os.File
	if *inputFile != "" {
		input, err = os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	out, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[!] Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	var log *os.File
	if *logFile != "" {
		log, err = os.Create(*logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] Error creating log: %v\n", err)
			os.Exit(1)
		}
		defer log.Close()
	}

	scanner := bufio.NewScanner(input)
	var wg sync.WaitGroup
	var mu sync.Mutex
	canaryIndex := 0
	lineNum := 0

	sem := make(chan struct{}, 20) // Thread pool size
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		sem <- struct{}{}
		wg.Add(1)
		go func(line string, ln int) {
			defer func() {
				<-sem
				wg.Done()
			}()
			outLine, logLine := injectToken(line, ln, &canaryIndex, &mu)
			if outLine != "" {
				mu.Lock()
				fmt.Fprintln(out, outLine)
				if log != nil {
					fmt.Fprintln(log, logLine)
				}
				if *verbose {
					fmt.Printf("[+] Injected: %s\n", outLine)
				}
				mu.Unlock()
			} else if *verbose {
				fmt.Printf("[-] Skipped: %s\n", line)
			}
		}(line, lineNum)
	}

	wg.Wait()
	if !*quiet {
		fmt.Printf("[+] Done. Output written to %s\n", *output)
		if *logFile != "" {
			fmt.Printf("[+] Log written to %s\n", *logFile)
		}
	}
}
