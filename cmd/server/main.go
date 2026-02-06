package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"posts-comments-1/graph"
	"posts-comments-1/internal/storage/memory"
	pg "posts-comments-1/internal/storage/postgres"
)

func main() {
	storageType := os.Getenv("STORAGE_TYPE")

	var resolver graph.Resolver
	if storageType == "postgres" {
		pgs, err := pg.New()
		if err != nil {
			log.Fatal(err)
		}
		resolver = graph.Resolver{Storage: pgs}
	} else {
		resolver = graph.Resolver{Storage: memory.New()}
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: &resolver,
	}))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
