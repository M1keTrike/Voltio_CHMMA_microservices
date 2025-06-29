// main is the entry point of the WebSocketServer application.
// It starts the server by invoking server.StartServer from the internal/server package.
package main

import "github.com/M1keTrike/EventDriven/internal/server"

func main() {
	server.StartServer()
}
