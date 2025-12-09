package main

// file should hold all feed handler functions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/MadridMAC/gator/internal/database"
	"github.com/google/uuid"
)

// usage: gator agg [time_between_reqs] | time_between_reqs is a duration string (e.g. 30s for 30 seconds, 1m for 1 minute)
// Fetches and parses the user's added feeds. Titles are printed to the console, while posts are saved to the database.
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

// usage: gator addfeed [feed_name] [url]
// Adds an RSS feed for the current user.
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
	log.Printf("successfully added new feed %s\n", feed_fields.Channel.Title)
	return nil
}

// usage: gator feeds
// Shows all registered feeds for all users
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

// usage: gator follow [url]
// Follow an existing feed registered by another user
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

// usage: gator following
// Shows the RSS feeds that the current/active user is following
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

// usage: gator unfollow [url]
// Unfollows a specific feed for the current user via URL
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

// usage: gator browse [limit]
// Browses the active user's currently saved posts, up to the provided limit. If no limit is provided, the default is 2.
func handlerBrowse(s *state, cmd command, curr_user database.User) error {
	if len(cmd.args) > 1 {
		log.Fatal("error: expecting zero or 1 Limit argument")
	}

	limit := 2

	if len(cmd.args) == 1 {
		new_limit, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			log.Fatalf("error parsing provided limit argument: %v", err)
		}

		limit = int(new_limit)
	}

	get_post_params := database.GetPostsForUserParams{
		UserID: curr_user.ID,
		Limit:  int32(limit),
	}

	post_output, err := s.db.GetPostsForUser(context.Background(), get_post_params)
	if err != nil {
		log.Fatalf("error fetching user's posts: %v\n", err)
	}

	fmt.Printf("%s's Feed Posts:\n", curr_user.Name)
	for _, item := range post_output {
		fmt.Printf("* Title: %v\n", item.Title)
		fmt.Printf("- URL: %v\n", item.Url)
		fmt.Printf("- Publish Date: %v\n", item.PublishedAt)
		fmt.Printf("- Description: %v\n", item.Description.String)
	}

	return nil
}

// Helper function for the agg command
// Scrapes the latest feeds and adds their posts to the database
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
	returned_feed, err := fetchFeed(context.Background(), latest_marked.Url)
	if err != nil {
		log.Fatalf("error fetching marked feed from URL: %v", err)
	}

	// second cleaning just in case
	fetched_feed := feedFormatter(returned_feed)

	fmt.Println("Saved Posts:")
	for _, item := range fetched_feed.Channel.Item {
		fmt.Printf("* %s\n", item.Title)

		_, err := s.db.GetPostFromUrl(context.Background(), fetched_feed.Channel.Link)
		if err == nil {
			continue
		}

		pubDate, err := time.Parse(time.RFC1123Z, string(item.PubDate))
		if err != nil {
			log.Fatalf("error parsing publish date: %v", err)
		}

		post_params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: pubDate,
			FeedID:      latest_marked.ID,
		}

		_, new_post_err := s.db.CreatePost(context.Background(), post_params)
		if new_post_err != nil {
			log.Fatalf("error creating post: %v", err)
		}
	}
	fmt.Println("--------------------------------")

	return nil
}
