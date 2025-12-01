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

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Fatal("error: no arguments found; expected single argument")
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

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		log.Fatal("error: no arguments found; expected single name argument")
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

func handlerReset(s *state, cmd command) error {
	del_users := s.db.DeleteUsers(context.Background())
	if del_users != nil {
		log.Fatalf("an error occurred while resetting the users table: %v\n", del_users)
	}
	return nil
}

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

func handlerAgg(s *state, cmd command) error {
	rss_feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		log.Fatalf("error retrieving RSS feed: %v", err)
	}
	cleaned_feed := feedFormatter(rss_feed)
	fmt.Println(cleaned_feed)
	return nil
}

func handlerAddFeed(s *state, cmd command, curr_user database.User) error {
	if len(cmd.args) < 2 {
		log.Fatal("error: insufficient arguments; expecting name and url argument")
	}
	// getting name and url
	name_arg := cmd.args[0]
	url_arg := cmd.args[1]

	// building params in feedData to create new feed
	feedData := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name_arg,
		Url:       url_arg,
		UserID:    curr_user.ID,
	}

	// create new feed w/ feedData
	new_feed, err := s.db.CreateFeed(context.Background(), feedData)
	if err != nil {
		log.Fatalf("error creating feed: %v", err)
	}

	// automatically follow new feed when created
	feed_follow_params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    curr_user.ID,
		FeedID:    new_feed.ID,
	}
	_, feed_follow_err := s.db.CreateFeedFollow(context.Background(), feed_follow_params)
	if feed_follow_err != nil {
		log.Fatalf("error creating feed follow: %v", err)
	}

	// fetch feed fields
	feed_fields, err := fetchFeed(context.Background(), new_feed.Url)
	if err != nil {
		log.Fatalf("error fetching feed: %v", err)
	}

	// success message + printing fields
	fmt.Printf("successfully created new feed %s\nfeed fields: %v\n", new_feed.Name, feed_fields)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	// get all feeds (name, url, attached user_id)
	all_feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		log.Fatalf("error fetching feeds: %v", err)
	}

	fmt.Println("List of Feeds:")
	for _, feed := range all_feeds {
		// fetch feed user's name via their user_id
		feed_user, err := s.db.GetFeedUser(context.Background(), feed.UserID)
		if err != nil {
			log.Fatalf("error fetching feed user: %v", err)
		}

		fmt.Printf("* %s | %s | from user: %s\n", feed.Name, feed.Url, feed_user)
	}
	return nil
}

func handlerFollow(s *state, cmd command, curr_user database.User) error {
	if len(cmd.args) == 0 {
		log.Fatal("error: expecting url argument")
	}

	feed_from_url, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		log.Fatal("error fetching feed from provided URL")
	}

	feed_follow_params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    curr_user.ID,
		FeedID:    feed_from_url.ID,
	}
	created_feed, err := s.db.CreateFeedFollow(context.Background(), feed_follow_params)
	if err != nil {
		log.Fatalf("error creating feed follow: %v", err)
	}

	fmt.Printf("current user %s successfully followed feed %s\n", created_feed.UserName, created_feed.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, curr_user database.User) error {
	following_list, err := s.db.GetFeedFollowsForUser(context.Background(), curr_user.ID)
	if err != nil {
		log.Fatal("error fetching following list")
	}

	fmt.Printf("user %s's follow list:\n", curr_user.Name)
	for _, followed_feed := range following_list {
		fmt.Printf("* %s\n", followed_feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, curr_user database.User) error {
	if len(cmd.args) == 0 {
		log.Fatal("error: expecting url argument")
	}

	feed_from_url, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		log.Fatal("error fetching feed from given url")
	}

	unfollow_params := database.DeleteFeedFollowParams{
		UserID: curr_user.ID,
		FeedID: feed_from_url.ID,
	}

	unfollow_feed := s.db.DeleteFeedFollow(context.Background(), unfollow_params)
	if unfollow_feed != nil {
		log.Fatalf("error unfollowing feed: %v", unfollow_feed)
	}

	return nil
}
