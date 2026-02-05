package main

import "posts-comments-1/internal/storage/postgres"

func main() {
	postgres.CheckConnection()
}
