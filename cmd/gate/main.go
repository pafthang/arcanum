// Command gate is the process entrypoint for the Optima HTTP edge.
//
//	go run ./cmd/gate
package main

import "github.com/pafthang/arcanum/services/gate"

func main() {
	gate.Run()
}
