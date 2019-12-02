# Advent of Code Downloader

[![Go Report Card](https://goreportcard.com/badge/github.com/GreenLightning/advent-of-code-downloader)](https://goreportcard.com/report/github.com/GreenLightning/advent-of-code-downloader)

`aocdl` is a command line utility that automatically downloads your [Advent of
Code](https://adventofcode.com/) puzzle inputs.

*Trivia*: If the puzzle input is very short, it will be embedded into the
puzzle page instead of being linked (for an example see [day 4 of
2015](https://adventofcode.com/2015/day/4)). Thanks to the consistent API of
the Advent of Code website, these puzzle inputs can be downloaded exactly like
the normal, longer puzzle inputs.

## Installation

Please use the standard `go get` command to build and install `aocdl`.

```
go get github.com/GreenLightning/advent-of-code-downloader/aocdl
```

## Setting the Session Cookie

Your session cookie is required to download your personalized puzzle input. You
can set it in two ways.

Provide your session cookie as a command line parameter:

```
aocdl -session-cookie 0123456789...abcdef
```

Or create a configuration file named `.aocdlconfig` in your home directory or in
the current directory and add the `session-cookie` key:

```json
{
	"session-cookie": "0123456789...abcdef"
}
```

## Basic Usage

Assuming you have created a configuration file (if not you must provide your
session cookie as a parameter), the following command will attempt to download
the input for the current day and save it to a file named `input.txt`:

```
aocdl
```

If you specify the `-wait` flag, the program will display a countdown waiting
for midnight (when new puzzles are released) and then download the input of
the new day:

```
aocdl -wait
```

## Options

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

## Configuration Files

The program looks for configuration files named `.aocdlconfig` in the user's
home directory and in the current working directory.

For each option, the configuration file in the current directory overwrites the
configuration file in the home directory and command line parameters overwrite
any configuration file.

Configuration files must contain one valid JSON object. The following keys
corresponding to some of the command line parameters above are accepted:

| Key              | Type   |
| ---------------- | ------ |
| "session-cookie" | String |
| "output"         | String |
| "year"           | Number |
| "day"            | Number |

A fully customized configuration file might look like this, although the program
would only ever download the same input unless the date is specified on the
command line:

```json
{
	"session-cookie": "0123456789...abcdef",
	"output": "input-{{.Year}}-{{.Day}}.txt",
	"year": 2015,
	"day": 24
}
```
