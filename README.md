# Hush: WebSSH Terminal Aggregator

Hush is a web-based interface for connecting to and managing multiple SSH terminals from your browser. It provides a modern, responsive UI for securely accessing remote servers, running commands, and viewing output in real time—all without leaving your web browser.

## Features

- **Web-based SSH Terminal:** Connect to any SSH server using a simple web form.
- **Real-time Terminal:** Interactive terminal experience powered by WebSockets.
- **Multiple Connections:** Easily switch between different servers (one at a time per session).
- **Modern UI:** Built with Tailwind CSS for a sleek, dark-themed terminal look.
- **Secure:** Credentials are not stored; connections are established per session.


## Getting Started

### Prerequisites
- Go 1.20+

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/hush.git
   cd hush
   ```
2. **Install dependencies:**
   ```bash
   go mod tidy
   ```
3. **Run the server:**
   ```bash
   go run main.go
   ```
4. **Open your browser:**
   Visit [http://localhost:8080](http://localhost:8080)

## Usage

1. Enter the SSH server details (host, port, username, password) in the form.
2. Click **Connect** to open a web-based terminal session.
3. Type commands in the terminal and see output in real time.
4. Use the **Clear** button to clear the terminal, or **Disconnect** to return to the form.

## Project Structure

```
├── main.go              # Main Go application (Gin, WebSocket, SSH logic)
├── go.mod, go.sum       # Go modules and dependencies
├── templates/
│   ├── form.html        # Connection form UI
│   └── terminal.html    # Terminal UI
└── README.md            # This file
```

## Dependencies
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket support
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) - SSH client
- [Tailwind CSS](https://tailwindcss.com/) - UI styling (via CDN)

## Security Notes
- Passwords are sent over the network; use HTTPS in production.
- Host key verification is disabled for demo purposes. For production, implement strict host key checking.

## License

MIT License. See [LICENSE](LICENSE) for details.
