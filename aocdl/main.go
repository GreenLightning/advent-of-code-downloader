package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	_ "time/tzdata"
)

const titleAboutMessage = `Advent of Code Downloader

aocdl is a command line utility that automatically downloads your Advent of Code
puzzle inputs.
`

const usageMessage = `Usage:

	aocdl [options]

Options:

	-session-cookie 0123456789...abcdef
		Use the specified string as session cookie.

	-output input.txt
		Save the downloaded puzzle input to the specified file. The special
		markers {{.Year}} and {{.Day}} will be replaced with the selected year
		and day. [see also Go documentation for text/template]

	-year 2000
	-day 24
		Download the input from the specified year or day. By default the
		current year and day is used.

	-force
		Overwrite file if it already exists.

	-wait
		If this flag is specified, year and day are ignored and the program
		waits until midnight (when new puzzles are released) and then downloads
		the input of the new day. While waiting a countdown is displayed. To
		reduce load on the Advent of Code servers, the download is started after
		a random delay between 2 and 30 seconds after midnight.
`

const repositoryMessage = `Repository:

	https://github.com/GreenLightning/advent-of-code-downloader
`

const missingSessionCookieMessage = `No Session Cookie

A session cookie is required to download your personalized puzzle input.

Please provide your session cookie as a command line parameter:

aocdl -session-cookie 0123456789...abcdef

Or create a configuration file named '.aocdlconfig' in your home directory or in
the current directory and add the 'session-cookie' key:

{
	"session-cookie": "0123456789...abcdef"
}
`

func main() {
	rand.Seed(time.Now().Unix())

	config, err := loadConfigs()
	checkError(err)

	addFlags(config)

	if config.SessionCookie == "" {
		fmt.Fprintln(os.Stderr, missingSessionCookieMessage)
		os.Exit(1)
	}

	est, err := time.LoadLocation("EST")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load time zone information:", err)
		os.Exit(1)
	}

	now := time.Now().In(est)
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, est)

	if config.Year == 0 {
		config.Year = now.Year()
	}
	if config.Day == 0 {
		config.Day = now.Day()
	}
	if config.Output == "" {
		config.Output = "input.txt"
	}

	if config.Wait {
		// Overwrite values before rendering output.
		config.Year = next.Year()
		config.Day = next.Day()
	}

	err = renderOutput(config)
	checkError(err)

	// Check if output file exists before waiting and before downloading.
	info, err := os.Stat(config.Output)
	if err == nil {
		if info.IsDir() {
			fmt.Fprintf(os.Stderr, "cannot write to '%s' because it is a directory\n", config.Output)
			os.Exit(1)
		}
		if !config.Force {
			fmt.Fprintf(os.Stderr, "file '%s' already exists; use '-force' to overwrite\n", config.Output)
			os.Exit(1)
		}
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "failed to check output file '%s': %v\n", config.Output, err)
		os.Exit(1)
	}

	if config.Wait {
		wait(next)
	}

	err = download(config)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addFlags(config *configuration) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)

	ignored := new(bytes.Buffer)
	flags.SetOutput(ignored)

	sessionCookieFlag := flags.String("session-cookie", "", "")
	outputFlag := flags.String("output", "", "")
	yearFlag := flags.String("year", "", "")
	dayFlag := flags.String("day", "", "")

	forceFlag := flags.Bool("force", false, "")
	waitFlag := flags.Bool("wait", false, "")

	var year, day int

	flagErr := flags.Parse(os.Args[1:])

	if flagErr == nil {
		year, flagErr = parseIntFlag(*yearFlag)
	}

	if flagErr == nil {
		day, flagErr = parseIntFlag(*dayFlag)
	}

	if flagErr == flag.ErrHelp {
		fmt.Println(titleAboutMessage)
		fmt.Println(usageMessage)
		fmt.Println(repositoryMessage)
		os.Exit(0)
	}

	if flagErr != nil {
		fmt.Fprintln(os.Stderr, flagErr)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(1)
	}

	flagConfig := new(configuration)
	flagConfig.SessionCookie = *sessionCookieFlag
	flagConfig.Output = *outputFlag
	flagConfig.Year = year
	flagConfig.Day = day

	config.merge(flagConfig)

	if *forceFlag {
		config.Force = true
	}
	if *waitFlag {
		config.Wait = true
	}
}

func parseIntFlag(text string) (int, error) {
	if text == "" {
		return 0, nil
	}
	// Parse in base 10.
	value, err := strconv.ParseInt(text, 10, 0)
	return int(value), err
}

func renderOutput(config *configuration) error {
	tmpl, err := template.New("output").Parse(config.Output)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	data := make(map[string]int)
	data["Year"] = config.Year
	data["Day"] = config.Day

	err = tmpl.Execute(buf, data)
	if err != nil {
		return err
	}

	config.Output = buf.String()

	return nil
}

func wait(next time.Time) {
	min, max := 2*1000, 30*1000
	delayMillis := min + rand.Intn(max-min+1)

	hours, mins, secs := 0, 0, 0
	for remaining := time.Until(next); remaining >= 0; remaining = time.Until(next) {
		remaining += 1 * time.Second // let casts round up instead of down
		newHours := int(remaining.Hours()) % 24
		newMins := int(remaining.Minutes()) % 60
		newSecs := int(remaining.Seconds()) % 60
		if newHours != hours || newMins != mins || newSecs != secs {
			hours, mins, secs = newHours, newMins, newSecs
			fmt.Printf("\r%02d:%02d:%02d + %04.1fs", hours, mins, secs, float32(delayMillis)/1000.0)
		}
		time.Sleep(200 * time.Millisecond)
	}

	next = next.Add(time.Duration(delayMillis) * time.Millisecond)

	millis := 0
	for remaining := time.Until(next); remaining >= 0; remaining = time.Until(next) {
		newMillis := int(remaining.Nanoseconds() / 1e6)
		if newMillis != millis {
			millis = newMillis
			fmt.Printf("\r00:00:00 + %04.1fs", float32(millis)/1000.0)
		}
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Printf("\r                \r")
}

func download(config *configuration) error {
	client := new(http.Client)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://adventofcode.com/%d/day/%d/input", config.Year, config.Day), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "github.com/GreenLightning/advent-of-code-downloader")

	cookie := new(http.Cookie)
	cookie.Name, cookie.Value = "session", config.SessionCookie
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if config.Force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(config.Output, flags, 0666)
	if os.IsExist(err) {
		return fmt.Errorf("file '%s' already exists; use '-force' to overwrite", config.Output)
	} else if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
