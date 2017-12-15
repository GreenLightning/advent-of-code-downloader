package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
	"math/rand"
	"net/http"
	"text/template"
)

const titleAboutMessage =
`Advent of Code Downloader

aocdl is a command line utility that automatically downloads your Advent of Code
puzzle inputs.
`

const usageMessage =
`Usage:

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

	-wait
		If this flag is specified, year and day are ignored and the program
		waits until midnight (when new puzzles are released) and then downloads
		the input of the new day. While waiting a countdown is displayed. To
		reduce load on the Advent of Code servers, the download is started after
		a random delay between 2 and 30 seconds after midnight.
`

const repositoryMessage =
`Repository:

	https://github.com/GreenLightning/advent-of-code-downloader
`

const missingSessionCookieMessage =
`No session cookie provided. The session cookie is required to download your
personalized puzzle input.

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
	checkError(err)

	now := time.Now().In(est)
	next := time.Date(now.Year(), now.Month(), now.Day() + 1, 0, 0, 0, 0, est)

	if config.Year == 0 { config.Year = now.Year() }
	if config.Day  == 0 { config.Day  = now.Day()  }
	if config.Output == "" { config.Output = "input.txt" }

	if config.Wait {
		// Overwrite values before rendering output.
		config.Year = next.Year()
		config.Day  = next.Day()
	}

	err = renderOutput(config)
	checkError(err)

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

	helpFlag := flags.Bool("help", false, "")

	sessionCookieFlag := flags.String("session-cookie", "", "")
	outputFlag := flags.String("output", "", "")
	yearFlag := flags.Int("year", 0, "")
	dayFlag := flags.Int("day", 0, "")

	waitFlag := flags.Bool("wait", false, "")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Println(titleAboutMessage)
		fmt.Println(usageMessage)
		fmt.Println(repositoryMessage)
		os.Exit(0)
	}

	flagConfig := new(configuration)
	flagConfig.SessionCookie = *sessionCookieFlag
	flagConfig.Output = *outputFlag
	flagConfig.Year = *yearFlag
	flagConfig.Day = *dayFlag

	config.merge(flagConfig)

	if *waitFlag { config.Wait = true }
}

func renderOutput(config *configuration) error {
	tmpl, err := template.New("output").Parse(config.Output)
	if err != nil { return err }

	buf := new(bytes.Buffer)

	data := make(map[string]int)
	data["Year"] = config.Year
	data["Day"] = config.Day

	err = tmpl.Execute(buf, data)
	if err != nil { return err }

	config.Output = buf.String()

	return nil
}

func wait(next time.Time) {
	min, max := 2 * 1000, 30 * 1000
	delayMillis := min + rand.Intn(max - min + 1)

	hours, mins, secs := 0, 0, 0
	for remaining := time.Until(next); remaining >= 0; remaining = time.Until(next) {
		remaining += 1 * time.Second // let casts round up instead of down
		newHours := int(remaining.Hours()) % 24
		newMins  := int(remaining.Minutes()) % 60
		newSecs  := int(remaining.Seconds()) % 60
		if newHours != hours || newMins != mins || newSecs != secs {
			hours, mins, secs = newHours, newMins, newSecs
			fmt.Printf("\r%02d:%02d:%02d + %04.1fs", hours, mins, secs, float32(delayMillis) / 1000.0)
		}
		time.Sleep(200 * time.Millisecond)
	}

	next = next.Add(time.Duration(delayMillis) * time.Millisecond)

	millis := 0
	for remaining := time.Until(next); remaining >= 0; remaining = time.Until(next) {
		newMillis := int(remaining.Nanoseconds() / 1e6)
		if newMillis != millis {
			millis = newMillis
			fmt.Printf("\r00:00:00 + %04.1fs", float32(millis) / 1000.0)
		}
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Printf("\r                \r")
}

func download(config *configuration) error {
	client := new(http.Client)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://adventofcode.com/%d/day/%d/input", config.Year, config.Day), nil)
	if err != nil { return err }

	cookie := new(http.Cookie)
	cookie.Name, cookie.Value = "session", config.SessionCookie
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil { return err }

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	file, err := os.Create(config.Output)
	if err != nil { return err }

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil { return err }

	return nil
}
