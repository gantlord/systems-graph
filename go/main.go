package main

import (
	"context"
	"fmt"
	"math/rand"

	driver "github.com/arangodb/go-driver"
	http "github.com/arangodb/go-driver/http"
)

var collectionSize = 1000

func main() {
	// Set up a connection to the database
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		panic(err)
	}

	// Replace "username" and "password" with your ArangoDB credentials
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("username", "password"),
	})
	if err != nil {
		panic(err)
	}

	db, err := client.Database(nil, "_system")

	coll, err := db.Collection(context.Background(), "physicalServers")
	if err != nil {
		panic(err)
	}

	// generate random documents and insert them into the collection
	for i := 0; i < collectionSize; i++ {
		cores := rand.Intn(32) // generate a random number of cores between 0 and 31
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("opamux%d", i),
			"cores": cores,
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			//panic(err)
		}
	}

	// Retrieve the collections you want to print documents from
	collectionNames := []string{"binaries", "firewallRules", "physicalServers", "edges", "abstractServers", "components", "people", "purpose"}

	// Iterate over the collections and print out all documents in each collection
	for _, collName := range collectionNames {
		// Set up a query to retrieve all documents in the collection
		ctx := context.Background()
		query := "FOR doc IN @@collection RETURN doc"
		bindVars := map[string]interface{}{
			"@collection": collName,
		}
		cursor, err := db.Query(ctx, query, bindVars)
		if err != nil {
			panic(err)
		}

		// Iterate over the cursor and print out each document
		var doc interface{}
		for {
			_, err := cursor.ReadDocument(ctx, &doc)
			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				panic(err)
			}
			fmt.Printf("%v\n", doc)
		}

	}
}

