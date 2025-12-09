# gator
Gator is a CLI RSS feed aggregator built in Go, supporting the collection of multiple RSS feeds, storage of posts in a PostgreSQL database, ability to follow other users' feeds, and the browsing of aggregated post summaries with a link to the full post for convenience.

This program was originally made for Boot.dev's Blog Aggregator in Go project.

## Prerequisites
This program requires [PostgreSQL](https://www.postgresql.org/) and [Go](https://go.dev/).

## Installation
### Basic Installation
After downloading the repository, you can install Gator by running `go install` (for global usage) or `go build` (for local/specific directories)

### Config Setup
Gator requires the creation of a `.gatorconfig.json` file in your **home directory**. This .json file should contain a `db_url` variable with the connection credentials for the PostgreSQL database and a `current_user_name` variable of the current logged-in user, as seen in the sample below. 
```
{
  "db_url": "connection_string_goes_here",
  "current_user_name": "username_goes_here"
}

```
### Syntax
After installation and setting up the `.gatorconfig.json` file, you should be able to run Gator through your preferred CLI with the following syntax:
```
gator [command] [args]
```

## Usage
### User Registration
Before running any other command in Gator, you should register yourself or the current user with:
```
gator register [preferred_username]
```
Running this command will create the new user in the database, set the config file to the current user, and allow you to use the rest of Gator's commands without issue.

### Command List
Below are some of the commands that Gator supports. You can check the program's other commands in `cli_handler.go` and `feed_handlers.go`.

- `gator register [username]` - Registers a new user into the database and logs in as that user.
- `gator login [username]` - Logs into another already-existing user in the database.
- `gator reset` - Resets the user database.
- `gator addfeed [feed_name] [url]` - Adds an RSS feed for the current user.
- `gator agg [time_between_reqs]` - Fetches posts from the user's feeds. Prints titles and saves posts to the database.
- `gator browse [limit]` - Browse the user's currently saved posts, up to the provided limit (default 2).