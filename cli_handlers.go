package main

// file should hold all handler functions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MadridMAC/gator/internal/database"
	"github.com/google/uuid"
)

// usage: gator login [username]
// Logs into another already-existing user in the database
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		log.Fatal("error: expecting single name argument")
	}

	// check if user exists, because you can't login to an account that doesn't exist
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		log.Fatalf("error: user %s does not exist", cmd.args[0])
	}

	s.pointer.SetUser(cmd.args[0])
	fmt.Printf("username successfully set to %s\n", cmd.args[0])
	return nil
}

// usage: gator register [username]
// Registers a new user into the database and logs in as that user
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		log.Fatal("error: expecting single name argument")
	}
	name_arg := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), name_arg)

	if err != sql.ErrNoRows {
		log.Fatalf("error: user with name %s already exists", name_arg)
	}

	userData := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name_arg,
	}

	s.db.CreateUser(context.Background(), userData)
	s.pointer.SetUser(name_arg)

	fmt.Printf("user %s successfully registered\n", name_arg)
	newUser, _ := s.db.GetUser(context.Background(), name_arg)
	fmt.Printf("debug info: %v\n", newUser)

	return nil
}

// usage: gator reset
// Resets the user table/database.
func handlerReset(s *state, cmd command) error {
	del_users := s.db.DeleteUsers(context.Background())
	if del_users != nil {
		log.Fatalf("an error occurred while resetting the users table: %v\n", del_users)
	}
	return nil
}

// usage: gator users
// Shows a list of all users in the database, including the current/active user.
func handlerUsers(s *state, cmd command) error {
	user_list, err := s.db.GetUsers(context.Background())
	if err != nil {
		log.Fatalf("an error occurred while getting all users: %v\n", err)
	}
	for _, user := range user_list {
		if strings.EqualFold(user, s.pointer.Current_user_name) {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}
