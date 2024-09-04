package main

import (
	"log"
	"net/http"

	"html/template"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

type Responder struct {
	upgrader websocket.Upgrader
	input    chan WsMessage
}

func (rsp Responder) echo(w http.ResponseWriter, r *http.Request) {
	c, err := rsp.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for msg := range rsp.input {
		err = c.WriteMessage(websocket.TextMessage, []byte(msg.Text))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
	/* for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	} */
}

/* func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
} */

func home(w http.ResponseWriter, r *http.Request) {
	fileList := CollectRelativeFilePaths("/usr/share/nginx/my-vue-pwa/")
	data := struct {
		WsUrl    string
		FileList []string
		NginxUrl string
	}{
		WsUrl:    "ws://" + r.Host + "/echo",
		FileList: fileList,
		NginxUrl: "http://localhost",
	}
	homeTemplate.Execute(w, data)
}

type WsMessage struct {
	Text string
}

func startWsServer(c chan WsMessage, addr string) {
	// http.HandleFunc("/echo", echo)
	responder := Responder{input: c, upgrader: upgrader}
	http.HandleFunc("/echo", responder.echo)

	fs := http.FileServer(http.Dir("./public/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<link href="/assets/styles.css" rel="stylesheet" />
<script type="text/javascript" src="/assets/app.js"></script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
<ul>
	{{range .FileList}}<li><button data-asset-file="{{.}}">GET</button> {{.}}</li>{{end}}
<ul>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
<script>
	initApp({wsUrl: "{{.WsUrl}}", nginxUrl: "{{.NginxUrl}}"})
</script>
</body>
</html>
`))
