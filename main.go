package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/MadridMAC/gator/internal/config"
	"github.com/MadridMAC/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	// load DB URL and connect to DB
	dbURL := "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("error opening connection to database")
	}
	dbQueries := database.New(db)

	configFile := config.Read()
	newState := state{
		pointer: &configFile,
		db:      dbQueries,
	}

	// commandlist init
	commandList := commands{
		commandMap: map[string]func(*state, command) error{},
	}
	// register commands here
	commandList.register("login", handlerLogin)
	commandList.register("register", handlerRegister)

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
