package main

import (
	"log"
	"os"

	"github.com/MadridMAC/gator/internal/config"
)

func main() {
	configFile := config.Read()
	newState := state{
		pointer: &configFile,
	}
	commandList := commands{
		commandMap: map[string]func(*state, command) error{},
	}
	commandList.register("login", handlerLogin)

	curr_args := os.Args
	if len(curr_args) < 2 {
		log.Fatal("error: less than 2 arguments")
	}

	curr_command := command{
		name: curr_args[1],
		args: curr_args[2:],
	}

	// **keeping for debugging**
	// fmt.Println(curr_command.name)
	// fmt.Println(curr_command.args)

	commandList.run(&newState, curr_command)

	// **keeping for debugging**
	// configFile.SetUser("madrid")
	// updatedConfig := config.Read()
	// fmt.Println(updatedConfig)
}
