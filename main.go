package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nxadm/tail"
)

var nginxErrorLog = GetEnvOrDefault("APP_NGINX_ERROR_LOG", "/var/log/nginx/error.log")
var addr = ":8080"
var nginxPortOnHost = GetEnvOrDefault("APP_NGINX_PORT_ON_HOST", "80")
var webPath = GetEnvOrDefault("APP_WEB_PATH", "/usr/share/nginx/my-vue-pwa")
var fetchViaProxy = GetEnvBool("APP_FETCH_VIA_PROXY", false)
var nginxHost = GetEnvOrDefault("APP_NGINX_HOST", "localhost")

func GetEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return boolValue
}

func main() {
	log.Println(
		"# Starting Tester App",
		"\n\tAPP_NGINX_ERROR_LOG: ", nginxErrorLog,
		"\n\tAPP_NGINX_PORT_ON_HOST: ", nginxPortOnHost,
		"\n\tAPP_NGINX_HOST: ", nginxHost,
		"\n\tAPP_WEB_PATH: ", webPath,
		"\n\tAPP_FETCH_VIA_PROXY: ", fetchViaProxy,
	)
	// Create a tail
	t, err := tail.TailFile(
		nginxErrorLog,
		tail.Config{Follow: true, ReOpen: true, CompleteLines: true, Location: &tail.SeekInfo{Whence: io.SeekEnd}},
		// tail.Config{Follow: false},
	)
	if err != nil {
		panic(err)
	}

	var logChannel chan WsMessage = make(chan WsMessage)

	var lineFilter LineFilter = LineFilter{filterList: []FilterItem{
		ByPrefix{Prefix: "http request line:"},
		ByPrefix{Prefix: "http uri:"},

		ByPrefix{Prefix: "test location:"},
		ByPrefix{Prefix: "using configuration "},

		ByPrefix{Prefix: "http script var:"},
		ByPrefix{Prefix: "trying to use file:"},
		ByPrefix{Prefix: "trying to use dir:"},
		ByPrefix{Prefix: "http filename:"},
		ByPrefix{Prefix: "http finalize request:"},
	}}

	var requestCounter int
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
				requestCounter += 1
			}
			item.push(parsedLogLine)
			/* if item.isComplete() {
				fmt.Printf("Sequence completed %s\n", item.PidAndTid)
				fmt.Printf("%#v", item)

				os.Exit(0)
			} */
			logSequenceMap[requestFullId] = item

			if item.isComplete() {
				jsonB, _ := json.Marshal(item.withFilteredLines(lineFilter))
				logChannel <- WsMessage{Text: string(jsonB)}
				delete(logSequenceMap, requestFullId)
			}

			// logChannel <- WsMessage{Text: parsedLogLine.Message}

			if isNewRequest {
				logChannel <- WsMessage{Text: fmt.Sprintf("\"Total requests: %d\"", requestCounter)}
			}
		}
	}()

	nginxPortOnHostInt, _ := strconv.Atoi(nginxPortOnHost)
	startWsServer(logChannel, addr, nginxPortOnHostInt, webPath, fetchViaProxy, nginxHost)

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
	Lines         []LogLine
	hasStartLine  bool
	hasEndLine    bool
}

func (s *LogSequence) push(line *LogLine) {
	s.Lines = append(s.Lines, *line)

	if strings.HasPrefix(line.Message, "http process request line") {
		s.hasStartLine = true
	}

	if strings.HasPrefix(line.Message, "http filename:") || strings.HasPrefix(line.Message, "http finalize request:") {
		s.hasEndLine = true
	}
}

func (s LogSequence) isComplete() bool {
	return s.hasEndLine && s.hasStartLine
}

func (s LogSequence) toWsMessageText() string {
	var lines []string = make([]string, len(s.Lines))

	/* var filter LineFilter = LineFilter{filterList: []FilterItem{
		ByPrefix{Prefix: "http request line:"},
		ByPrefix{Prefix: "http uri:"},

		ByPrefix{Prefix: "test location:"},
		ByPrefix{Prefix: "using configuration "},

		ByPrefix{Prefix: "http script var:"},
		ByPrefix{Prefix: "trying to use file:"},
		ByPrefix{Prefix: "http filename:"},
	}} */

	// for _, item := range filter.Filter(s.Lines) {
	for _, item := range s.Lines {
		lines = append(lines, item.Message)
	}

	return strings.Join(lines, "\n")
}

func (s LogSequence) withFilteredLines(lf LineFilter) *LogSequence {
	return &LogSequence{
		RequestFullId: s.RequestFullId,
		Lines:         lf.Filter(s.Lines),
	}
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
