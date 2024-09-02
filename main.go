package main

import (
	"fmt"
	"regexp"

	"github.com/nxadm/tail"
)

func main() {
	// Create a tail
	t, err := tail.TailFile(
		"/var/log/nginx/error.log", tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		panic(err)
	}

	// Print the text of each received line
	for line := range t.Lines {
		// fmt.Println(line.Text)

		parsedLogLine, error := parseLogLine(line.Text)
		if error != nil {
			fmt.Printf("ERROR: %s\n", error)
			continue
		}

		fmt.Printf("%#v\n", parsedLogLine)
	}
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
	LogLevelError LogLevel = "error"
)

type LogLine struct {
	LogLevel LogLevel
	Message  string
}

func parseLogLine(text string) (*LogLine, error) {
	// re := regexp.MustCompile(`(?P<Date>\d{4}/\d{2}/\d{2}) (?P<Time>\d{2}:\d{2}:\d{2}) \[(?P<Level>\w+)\] \d+#\d+: \*\d+ (?P<Message>.*?), client: (?P<ClientIP>[\d\.]+), server: (?P<ServerName>[^,]+), request: "(?P<Request>[^"]+)", host: "(?P<Host>[^"]+)"`)
	re := regexp.MustCompile(`(?P<Date>\d{4}/\d{2}/\d{2}) (?P<Time>\d{2}:\d{2}:\d{2}) \[(?P<Level>\w+)\] \d+#\d+: (?P<Message>.*)`)

	match := re.FindStringSubmatch(text)
	if len(match) == 0 {
		return &LogLine{}, fmt.Errorf("can not parse log string: %s", text)
	}

	groupNames := re.SubexpNames()

	result := &LogLine{}
	for i, name := range groupNames {
		if i != 0 && name != "" {
			// if i >= len(match) {
			// 	fmt.Errorf("Out of bounds: %s", text)
			// }
			// fmt.Printf("%s: %s\n", name, match[i])
			switch name {
			case "Level":
				result.LogLevel = LogLevel(match[i])
			case "Message":
				result.Message = match[i]
			}
		}
	}

	return result, nil
}
