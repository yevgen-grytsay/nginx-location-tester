package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nxadm/tail"
)

func main() {
	// Create a tail
	t, err := tail.TailFile(
		"/var/log/nginx/error.log",
		tail.Config{Follow: true, ReOpen: true},
		// tail.Config{Follow: false},
	)
	if err != nil {
		panic(err)
	}

	var logChannel chan WsMessage = make(chan WsMessage)

	go func() {
		// for line := range t.Lines {
		// 	logChannel <- WsMessage{Text: line.Text}
		// }
		var logSequenceMap = make(map[string]LogSequence)
		// var lastLogLine *LogLine

		// Print the text of each received line
		for line := range t.Lines {
			// fmt.Println(line.Text)

			parsedLogLine, error := parseLogLine(line.Text)
			if error != nil {
				// fmt.Printf("ERROR: %s\n", error)
				continue
			}

			// lastLogLine = parsedLogLine
			fmt.Printf("%#v\n", parsedLogLine)

			requestFullId := parsedLogLine.RequestFullId.id()
			item, ok := logSequenceMap[requestFullId]
			isNewRequest := false
			if !ok {
				item = LogSequence{RequestFullId: parsedLogLine.RequestFullId}
				isNewRequest = true
			}
			item.push(parsedLogLine)
			/* if item.isComplete() {
				fmt.Printf("Sequence completed %s\n", item.PidAndTid)
				fmt.Printf("%#v", item)

				os.Exit(0)
			} */
			logSequenceMap[requestFullId] = item

			if item.isComplete() {
				logChannel <- WsMessage{Text: item.toWsMessageText()}
			}

			// logChannel <- WsMessage{Text: parsedLogLine.Message}

			if isNewRequest {
				logChannel <- WsMessage{Text: fmt.Sprintf("Total requests: %d", len(logSequenceMap))}
			}
		}
	}()

	startWsServer(logChannel)

	/* for key, value := range logSequenceMap {
		fmt.Println("Key:", key, "Hash:", value.PidAndTid, "_", value.RequestId, "IsComplete:", fmt.Sprintf("%#v", value.isComplete()))
	} */
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
	LogLevelError LogLevel = "error"
)

type LogLine struct {
	LogLevel      LogLevel
	Message       string
	RequestFullId RequestFullId
}

type RequestFullId struct {
	PidAndTid string
	RequestId string
}

func (requestFullId RequestFullId) id() string {
	return fmt.Sprintf("%s_%s", requestFullId.PidAndTid, requestFullId.RequestId)
}

type LogSequence struct {
	RequestFullId RequestFullId
	lines         []LogLine
	hasStartLine  bool
	hasEndLine    bool
}

func (s *LogSequence) push(line *LogLine) {
	s.lines = append(s.lines, *line)

	if strings.HasPrefix(line.Message, "http process request line") {
		s.hasStartLine = true
	}

	if strings.HasPrefix(line.Message, "http filename:") {
		s.hasEndLine = true
	}
}

func (s LogSequence) isComplete() bool {
	return s.hasEndLine && s.hasStartLine
}

func (s LogSequence) toWsMessageText() string {
	var lines []string = make([]string, len(s.lines))

	for _, item := range s.lines {
		lines = append(lines, item.Message)
	}

	return strings.Join(lines, "\n")
	/* var sb strings.Builder

	for item := range s.lines {
		sb.WriteString(item.Message)
	} */
}

func parseLogLine(text string) (*LogLine, error) {
	// re := regexp.MustCompile(`(?P<Date>\d{4}/\d{2}/\d{2}) (?P<Time>\d{2}:\d{2}:\d{2}) \[(?P<Level>\w+)\] \d+#\d+: \*\d+ (?P<Message>.*?), client: (?P<ClientIP>[\d\.]+), server: (?P<ServerName>[^,]+), request: "(?P<Request>[^"]+)", host: "(?P<Host>[^"]+)"`)
	re := regexp.MustCompile(`(?P<Date>\d{4}/\d{2}/\d{2}) (?P<Time>\d{2}:\d{2}:\d{2}) \[(?P<Level>\w+)\] (?P<PidAndTid>\d+#\d+): \*(?P<RequestID>\d+) (?P<Message>.*)`)

	match := re.FindStringSubmatch(text)
	if len(match) == 0 {
		return &LogLine{}, fmt.Errorf("can not parse log string: %s", text)
	}

	groupNames := re.SubexpNames()

	result := &LogLine{}
	rid := RequestFullId{}
	for i, name := range groupNames {
		if i != 0 && name != "" {
			switch name {
			case "Level":
				result.LogLevel = LogLevel(match[i])
			case "Message":
				result.Message = match[i]
			case "PidAndTid":
				rid.PidAndTid = match[i]
			case "RequestID":
				rid.RequestId = match[i]
			}
		}
	}

	result.RequestFullId = rid

	return result, nil
}
