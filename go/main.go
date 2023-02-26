package main

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func main() {
	// Set up a connection to the ArangoDB instance.
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		log.Fatalf("Failed to set up connection to ArangoDB: %v", err)
	}

	// Set up authentication for the connection.
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("username", "password"),
	})
	if err != nil {
		log.Fatalf("Failed to set up authentication for ArangoDB connection: %v", err)
	}

	// Get a database from the ArangoDB client.
	db, err := client.Database(nil, "_system")
	if err != nil {
		log.Fatalf("Failed to get database from ArangoDB client: %v", err)
	}
	
	// List the collections in the ArangoDB database.
	ctx := context.Background()
	collections, err := db.Collections(ctx)
	if err != nil {
		log.Fatalf("Failed to list collections in ArangoDB database: %v", err)
	}

	for _, collection := range collections {
		fmt.Printf("Collection: %s\n", collection.Name())
	}
}

