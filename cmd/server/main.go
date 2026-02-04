package main

import "posts_comments_1/internal/storage/postgres"

func main() {
	postgres.CheckConnection()
}
