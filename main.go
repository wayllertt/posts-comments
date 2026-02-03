package main

import "posts-comments-1/internal/postgres_connection"

func main() {
	postgres_connection.CheckConnection()
}
