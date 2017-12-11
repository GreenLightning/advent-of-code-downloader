package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
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
	config, err := loadConfigs()
	checkError(err)

	addFlags(config)

	if config.SessionCookie == "" {
		fmt.Fprintln(os.Stderr, missingSessionCookieMessage)
		os.Exit(1)
	}

	err = addDefaultValues(config)
	checkError(err)

	err = renderOutput(config)
	checkError(err)

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

	err := flags.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Fprintln(os.Stderr, titleAboutMessage)
		fmt.Fprintln(os.Stderr, usageMessage)
		fmt.Fprintln(os.Stderr, repositoryMessage)
		os.Exit(0)
	}

	if *sessionCookieFlag != "" { config.SessionCookie = *sessionCookieFlag }
	if *outputFlag        != "" { config.Output        = *outputFlag        }
	if *yearFlag != 0 { config.Year = *yearFlag }
	if *dayFlag  != 0 { config.Day  = *dayFlag  }
}

func addDefaultValues(config *configuration) error {
	est, err := time.LoadLocation("EST")
	if err != nil { return err }

	now := time.Now().In(est)
	if config.Year == 0 { config.Year = now.Year() }
	if config.Day  == 0 { config.Day  = now.Day()  }

	if config.Output == "" { config.Output = "input.txt" }

	return nil
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
