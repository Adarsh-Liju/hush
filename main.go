package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type ConnInfo struct {
	Host     string
	Port     string
	User     string
	Password string
}

var currentConn ConnInfo
var mu sync.Mutex

var upgrader = websocket.Upgrader{}

func main() {
	http.HandleFunc("/", serveForm)
	http.HandleFunc("/connect", connectHandler)
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "form.html", nil)
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	host := r.FormValue("host")
	port := r.FormValue("port")
	if port == "" {
		port = "22"
	}

	user := r.FormValue("user")
	password := r.FormValue("password")

	mu.Lock()
	currentConn = ConnInfo{Host: host, Port: port, User: user, Password: password}
	mu.Unlock()

	renderTemplate(w, "terminal.html", currentConn)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	connInfo := currentConn
	mu.Unlock()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade failed:", err)
		return
	}
	defer ws.Close()

	config := &ssh.ClientConfig{
		User: connInfo.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(connInfo.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", connInfo.Host+":"+connInfo.Port, config)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("SSH connect failed: "+err.Error()))
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("SSH session failed: "+err.Error()))
		return
	}
	defer session.Close()

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	session.Stderr = session.Stdout
	session.Shell()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			ws.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		stdin.Write(msg)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles("templates/" + tmpl))
	t.Execute(w, data)
}
