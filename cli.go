package main

import (
	"fmt"

	"github.com/MadridMAC/gator/internal/config"
	"github.com/MadridMAC/gator/internal/database"
)

type state struct {
	pointer *config.Config
	db      *database.Queries
}

type command struct {
	name string
	args []string
}

// holds all CLI commands
type commands struct {
	commandMap map[string]func(*state, command) error
}

// runs a given command with the provided state if it exists
func (c *commands) run(s *state, cmd command) error {
	validCommand, ok := c.commandMap[cmd.name]
	if !ok {
		return fmt.Errorf("error: command does not exist")
	}
	return validCommand(s, cmd)
}

// registers a new handler function for a command name
func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}
