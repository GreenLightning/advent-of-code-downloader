# Advent of Code Downloader

[![Go Report Card](https://goreportcard.com/badge/github.com/GreenLightning/advent-of-code-downloader)](https://goreportcard.com/report/github.com/GreenLightning/advent-of-code-downloader)

`aocdl` is a command line utility that automatically downloads your [Advent of
Code](https://adventofcode.com/) puzzle inputs.

This tool is for competitive programmers, who want to solve the puzzles as
fast as possible. Using the `-wait` flag you can actually start the program
before the puzzle is published, thus spending exactly zero seconds of
competition time on downloading the input (see [Basic Usage](#basic-usage) for
details).

If you are working with the command line, it might also be more comfortable to
type `aocdl -year 2015 -day 1` instead of downloading the puzzle input using
the browser.

*Trivia*: If the puzzle input is very short, it will be embedded into the
puzzle page instead of being linked (for an example see [day 4 of
2015](https://adventofcode.com/2015/day/4)). Thanks to the consistent API of
the Advent of Code website, these puzzle inputs can be downloaded exactly like
the normal, longer puzzle inputs.

## Installation

#### Binary Download

You can download pre-compiled binaries from the
[releases](https://github.com/GreenLightning/advent-of-code-downloader/releases/latest/)
page. Just unzip the archive and place the binary in your working directory or
in a convenient location in your PATH.

#### Build From Source

If you have the [Go](https://golang.org/) compiler installed, you can use the
standard `go install` command to download, build and install `aocdl`.

```
go install github.com/GreenLightning/advent-of-code-downloader/aocdl@latest
```

## Setting the Session Cookie

Your session cookie is required to download your personalized puzzle input.
See the two sections below, if you want to know what a session cookie is or
how to get yours. The session cookies from the Advent of Code website are
valid for about a month, so you only have to get your cookie once per event.
You can provide it to `aocdl` in two ways.

Set your session cookie as a command line parameter:

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

You can test your setup by downloading an old puzzle input using `aocdl -year
2020 -day 1`. You will get an appropriate error message if the program cannot
find a session cookie. However, if you get a `500 Internal Server Error`, this
most likely means your session cookie is invalid or expired.

#### What Is a Session Cookie?

A session cookie is a small piece of data used to authenticate yourself to the
Advent of Code web servers. It is not human-readable and might look something
like this (this is not a valid cookie):

```
53616c7465645f5fbd2d445187c5dc5463efb7020021c273c3d604b5946f9e87e2dc30b649f9b2235e8cd57632e415cb
```

When you log in, the Advent of Code server generates a new session cookie and
sends it to your browser, which saves it on your computer. Every time you make
a request, your browser sends the cookie back to the server, which is how the
server knows that the request is from you and not somebody else. That way the
server can send you a personalized version of the website (for example
displaying your username and current number of stars or sending you your
personal puzzle input instead of somebody else's input).

#### How Do I Get My Session Cookie?

Google Chrome:

- Go to [adventofcode.com](https://adventofcode.com/)
- Make sure you are logged in
- Right click and select "Inspect"
- Select the "Application" tab
- In the tree on the left, select "Storage" → "Cookies" → "https://adventofcode.com"
- You should see a table of cookies, find the row with "session" as name
- Double click the row in the "Value" column to select the value of the cookie
- Press `CTRL + C` or right click and select "Copy" to copy the cookie
- Paste it into your configuration file or on the command line

Mozilla Firefox:

- Go to [adventofcode.com](https://adventofcode.com/)
- Make sure you are logged in
- Right click and select "Inspect Element"
- Select the "Storage" tab
- In the tree on the left, select "Cookies" → "https://adventofcode.com"
- You should see a table of cookies, find the row with "session" as name
- Double click the row in the "Value" column to select the value of the cookie
- Press `CTRL + C` or right click and select "Copy" to copy the cookie
- Paste it into your configuration file or on the command line

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

Finally, you can also specify a day (and year) explicitly.

```
aocdl -day 1
aocdl -year 2015 -day 1
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
