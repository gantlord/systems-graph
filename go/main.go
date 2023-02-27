package main

import (
	"context"
	"fmt"
	"math/rand"
	"test/utilities"

	driver "github.com/arangodb/go-driver"
	http "github.com/arangodb/go-driver/http"
)

var smallCollectionSize = 2
var mediumCollectionSize = 4
var largeCollectionSize = 6

func main() {
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

	collectionNames := []string{"binaries", "firewallRules", "physicalServers", "edges", "abstractServers", "components", "people", "purposes"}

	for _, collName := range collectionNames {

		if coll, err := db.Collection(nil, collName); err == nil {
			coll.Remove(nil)
		}

		ctx := context.Background()
		options := &driver.CreateCollectionOptions{}
		_, err := db.CreateCollection(ctx, collName, options)
		if err != nil {
			panic(err)
		}
	}

	coll, err := db.Collection(context.Background(), "binaries")
	if err != nil {
		panic(err)
	}

	for i := 0; i < smallCollectionSize; i++ {
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("binary%d", i),
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}

	coll, err = db.Collection(context.Background(), "firewallRules")
	if err != nil {
		panic(err)
	}

	for i := 0; i < smallCollectionSize; i++ {
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("rule%d", i),
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}

	coll, err = db.Collection(context.Background(), "abstractServers")
	if err != nil {
		panic(err)
	}

	for i := 0; i < smallCollectionSize; i++ {
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("abstractServers%d", i),
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}

	coll, err = db.Collection(context.Background(), "components")
	if err != nil {
		panic(err)
	}

	for i := 0; i < largeCollectionSize; i++ {
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("component%d", i),
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}

	coll, err = db.Collection(context.Background(), "purposes")
	if err != nil {
		panic(err)
	}

	for i := 0; i < smallCollectionSize; i++ {
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("purpose%d", i),
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}

	coll, err = db.Collection(context.Background(), "people")
	if err != nil {
		panic(err)
	}

	for i := 0; i < smallCollectionSize; i++ {
		doc := map[string]interface{}{
			"_key":  utilities.GetRandomName(),
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}

	coll, err = db.Collection(context.Background(), "physicalServers")
	if err != nil {
		panic(err)
	}

	for i := 0; i < mediumCollectionSize; i++ {
		cores := rand.Intn(32) 
		doc := map[string]interface{}{
			"_key":  fmt.Sprintf("server%d", i),
			"cores": cores,
		}
		_, err := coll.CreateDocument(context.Background(), doc)
		if err != nil {
			panic(err)
		}
	}


	for _, collName := range collectionNames {
		ctx := context.Background()
		query := "FOR doc IN @@collection RETURN doc"
		bindVars := map[string]interface{}{
			"@collection": collName,
		}
		cursor, err := db.Query(ctx, query, bindVars)
		if err != nil {
			panic(err)
		}

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

