// Command nats is the process entrypoint for the embedded NATS server.
//
//	go run ./cmd/nats
//
// Prefer repo root: go run ./cmd/nats
package main

import "github.com/pafthang/arcanum/services/nats"

func main() {
	nats.Run()
}
