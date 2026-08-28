// Command ctrl is the process entrypoint for the control plane.
//
//	go run ./cmd/ctrl
//	go run ./cmd/ctrl -up   # supervise full local stack
package main

import "github.com/pafthang/arcanum/services/ctrl"

func main() {
	ctrl.Run()
}
