package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 2 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 2 * time.Second
)

var (
	addr      = flag.String("addr", ":8080", "http service address")
	homeTempl = template.Must(template.New("").Parse(homeHTML))
	// filename  string
	basepath string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type FileInfo struct {
	LastMod  time.Time
	Options  []byte
	Filename string
}

func readFileIfModified(lastMod time.Time, filename string) (FileInfo, error) {
	fi, err := os.Stat(path.Join(basepath, filename))
	if err != nil {
		return FileInfo{lastMod, nil, filename}, err
	}
	if !fi.ModTime().After(lastMod) {
		return FileInfo{lastMod, nil, filename}, nil
	}
	p, err := ioutil.ReadFile(filename)
	if err != nil {
		return FileInfo{lastMod, nil, filename}, err
	}
	return FileInfo{fi.ModTime(), p, filename}, nil
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func writer(ws *websocket.Conn, lastMod time.Time, Filename string) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-fileTicker.C:
			var err error

			info, err := readFileIfModified(lastMod, Filename)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					info.Options = []byte(lastError)
				}
				log.Println(err)
			} else {
				lastError = ""
			}

			if info.Options != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if ret, err := json.Marshal(info); err != nil {
					log.Println(err)
					return
				} else {
					if err := ws.WriteMessage(websocket.TextMessage, ret); err != nil {
						log.Println(err)
						return
					}
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		log.Println(err)
		return
	}

	//var Filename string
	//if n, err := strconv.ParseInt(r.FormValue("filename"), 16, 64); err == nil {
	//	lastMod = time.Unix(0, n)
	//} else {
	//	log.Println(err)
	//}
	Filename := r.FormValue("filename")

	go writer(ws, time.Unix(0, 0), Filename)
	reader(ws)
}

//type UpdateBody struct {
//	Payload string
//}

func serveUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var p FileInfo
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//sDec, err := base64.StdEncoding.DecodeString(p.Payload)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//ret, err := base64.StdEncoding.DecodeString(p.Options)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}

	ioutil.WriteFile(p.Filename, p.Options, 0644)
	err = os.Chtimes(p.Filename, p.LastMod, p.LastMod)
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte("{}"))
}

func serveCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var p FileInfo
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := os.Stat(path.Join(basepath, p.Filename)); err == nil {
		http.Error(w, "已存在", http.StatusBadRequest)
		return
	}
	bytesRead, err := ioutil.ReadFile("tp")
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ioutil.WriteFile(p.Filename, bytesRead, 0644)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = os.Chtimes(p.Filename, p.LastMod, p.LastMod)
	if err != nil {
		fmt.Println(err)
	}
	w.Write([]byte("{}"))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	info, err := readFileIfModified(time.Time{}, filename)
	if err != nil {
		info.Options = []byte(err.Error())
		info.LastMod = time.Unix(0, 0)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	//tst := string(p)
	//if strings.Contains(tst, "\t") {
	//	fmt.Println("cont")
	//}
	//fmt.Println(tst)
	//var v = struct {
	//	Host    string
	//	Data    string
	//	LastMod string
	//}{
	//	r.Host,
	//	string(p),
	//	strconv.FormatInt(lastMod.UnixNano(), 16),
	//}
	//homeTempl.Execute(w, &v)
	// fmt.Println(strconv.FormatInt(info.LastMod.UnixNano(), 16))

	if ret, err := json.Marshal(info); err != nil {
		fmt.Println(err)
		http.Error(w, "Not found", http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(ret)
	}
	//dst := base64.StdEncoding.EncodeToString(p)
	//var data interface{}
	//err = json.Unmarshal(p, &data)
	//if err != nil {
	//	fmt.Println(err)
	//	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//	return
	//}
	//buffer := &bytes.Buffer{}
	//encoder := json.NewEncoder(buffer)
	//encoder.SetEscapeHTML(true)
	//err = encoder.Encode(p)
	//if err != nil {
	//	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//	return
	//}
	//resultJson, _ := gabs.Consume(p)
	//json.NewEncoder(w).Encode(data)
}

func main() {
	//flag.Parse()
	//if flag.NArg() != 1 {
	//	log.Fatal("filename not specified")
	//}
	// filename = flag.Args()[0]
	basepath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/create", serveCreate)
	mux.HandleFunc("/update", serveUpdate)
	mux.HandleFunc("/ws", serveWs)
	handler := cors.Default().Handler(mux)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}

const homeHTML = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>WebSocket Example</title>
    </head>
    <body>
        <pre id="fileData">{{.Data}}</pre>
        <script type="text/javascript">
            (function() {
                var data = document.getElementById("fileData");
                var conn = new WebSocket("ws://{{.Host}}/ws?lastMod={{.LastMod}}");
                conn.onclose = function(evt) {
                    data.textContent = 'Connection closed';
                }
                conn.onmessage = function(evt) {
                    console.log('file updated');
                    data.textContent = evt.data;
                }
            })();
        </script>
    </body>
</html>
`
