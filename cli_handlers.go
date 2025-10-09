package main

// file should hold all handler functions

import (
	"fmt"
	"log"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Fatal("error: no arguments found; expected single argument")
	}
	s.pointer.SetUser(cmd.args[0])
	fmt.Printf("username successfully set to %s", cmd.args[0])
	return nil
}
