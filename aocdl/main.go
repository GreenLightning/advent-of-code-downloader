package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
	"net/http"
	"text/template"
)

const sessionHelp =
`No session cookie provided. The session cookie is required to download your
personalized puzzle input.

Please create a configuration file named '.aocdlconfig' in your home directory
or in the current directory and add the 'session-cookie' key.

{
	"session-cookie": "0123456789...abcdef"
}
`

func main() {
	config, err := loadConfigs()
	checkError(err)

	if config.SessionCookie == "" {
		fmt.Fprintln(os.Stderr, sessionHelp)
		os.Exit(1)
	}

	err = addDefaultValues(config)
	checkError(err)

	err = renderOutput(config)
	checkError(err)

	client := new(http.Client)
	err, _ = download(config, client)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

	var buf bytes.Buffer

	data := make(map[string]int)
	data["Year"] = config.Year
	data["Day"] = config.Day

	err = tmpl.Execute(&buf, data)
	if err != nil { return err }

	config.Output = buf.String()

	return nil
}

func download(config *configuration, client *http.Client) (error, int) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://adventofcode.com/%d/day/%d/input", config.Year, config.Day), nil)
	if err != nil { return err, 0 }

	cookie := new(http.Cookie)
	cookie.Name, cookie.Value = "session", config.SessionCookie
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil { return err, 0 }

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status), resp.StatusCode
	}

	file, err := os.Create(config.Output)
	if err != nil { return err, 0 }

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil { return err, 0 }

	return nil, 0
}
