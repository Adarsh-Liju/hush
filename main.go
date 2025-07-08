package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type ConnInfo struct {
	Host     string `form:"host" json:"host"`
	Port     string `form:"port" json:"port"`
	User     string `form:"user" json:"user"`
	Password string `form:"password" json:"password"`
}

var currentConn ConnInfo
var mu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Parse templates once at startup
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Load HTML templates
	r.LoadHTMLGlob("templates/*.html")

	// Routes
	r.GET("/", serveForm)
	r.POST("/connect", connectHandler)
	r.GET("/ws", wsHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}

func serveForm(c *gin.Context) {
	c.HTML(http.StatusOK, "form.html", gin.H{})
}

func connectHandler(c *gin.Context) {
	var connInfo ConnInfo

	// Bind form data
	if err := c.ShouldBind(&connInfo); err != nil {
		log.Printf("Form binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Debug: log the received data
	log.Printf("Received connection info: Host=%s, Port=%s, User=%s", connInfo.Host, connInfo.Port, connInfo.User)

	// Set default port if not provided
	if connInfo.Port == "" {
		connInfo.Port = "22"
	}

	// Validate required fields
	if connInfo.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host is required"})
		return
	}
	if connInfo.User == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Store connection info
	mu.Lock()
	currentConn = connInfo
	mu.Unlock()

	// Render terminal page
	c.HTML(http.StatusOK, "terminal.html", connInfo)
}

func wsHandler(c *gin.Context) {
	mu.Lock()
	connInfo := currentConn
	mu.Unlock()

	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer ws.Close()

	// SSH configuration
	config := &ssh.ClientConfig{
		User: connInfo.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(connInfo.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH server
	sshClient, err := ssh.Dial("tcp", connInfo.Host+":"+connInfo.Port, config)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("SSH connect failed: "+err.Error()))
		return
	}
	defer sshClient.Close()

	// Create SSH session
	session, err := sshClient.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("SSH session failed: "+err.Error()))
		return
	}
	defer session.Close()

	// Set up pipes
	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	session.Stderr = session.Stdout

	// Start shell
	if err := session.Shell(); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("Failed to start shell: "+err.Error()))
		return
	}

	// Read from SSH stdout and send to WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				break
			}
		}
	}()

	// Read from WebSocket and send to SSH stdin
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		if _, err := stdin.Write(msg); err != nil {
			break
		}
	}
}
