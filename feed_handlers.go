package main

// file should hold all feed handler functions

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MadridMAC/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		log.Fatal("error: insufficient arguments; expecting single time_between_reqs string argument")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		log.Fatal("error: parsing time_between_reqs argument -- please check syntax")
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

	// rss_feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	// if err != nil {
	// 	log.Fatalf("error retrieving RSS feed: %v", err)
	// }
	// cleaned_feed := feedFormatter(rss_feed)
	// fmt.Println(cleaned_feed)
}

func handlerAddFeed(s *state, cmd command, curr_user database.User) error {
	if len(cmd.args) != 2 {
		log.Fatal("error: argument error; expecting name and url argument")
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
	log.Printf("successfully created new feed %s\nfeed fields: %v\n", new_feed.Name, feed_fields)
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
	if len(cmd.args) != 1 {
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
	if len(cmd.args) != 1 {
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

func scrapeFeeds(s *state) error {
	// fetch next feed from DB
	feed_from_db, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Fatalf("error fetching latest feed: %v", err)
	}

	// mark latest feed as fetched
	latest_marked, err := s.db.MarkFeedFetched(context.Background(), feed_from_db.ID)
	if err != nil {
		log.Fatalf("error marking latest feed: %v", err)
	}

	// fetch the marked feed via URL
	fetched_feed, err := fetchFeed(context.Background(), latest_marked.Url)
	if err != nil {
		log.Fatalf("error fetching marked feed from URL: %v", err)
	}

	fmt.Println("Fetched Posts:")
	for _, item := range fetched_feed.Channel.Item {
		fmt.Printf("* %s\n", item.Title)
	}

	return nil
}
