package main

import (
	"context"
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
	commandList.register("reset", handlerReset)
	commandList.register("users", handlerUsers)
	commandList.register("agg", handlerAgg)
	commandList.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commandList.register("feeds", handlerFeeds)
	commandList.register("follow", middlewareLoggedIn(handlerFollow))
	commandList.register("following", middlewareLoggedIn(handlerFollowing))
	commandList.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commandList.register("browse", middlewareLoggedIn(handlerBrowse))

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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		curr_user, err := s.db.GetUser(context.Background(), s.pointer.Current_user_name)
		if err != nil {
			log.Fatal("error fetching current user")
		}
		return handler(s, cmd, curr_user)
	}
}
