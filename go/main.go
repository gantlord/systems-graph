package main

import (
    "context"
    "fmt"
    "math/rand"
    "test/utilities"

    driver "github.com/arangodb/go-driver"
    http "github.com/arangodb/go-driver/http"
)

type CollectionInfo struct {
    Name     string
    Size     int
    DocGenFn func(i int) map[string]interface{}
}

var collections = []CollectionInfo{
    {"binaries", 1000,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("binary%d", i)} }},
    {"firewallRules", 100,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("rule%d", i)} }},
    {"abstractServers", 1000,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("abstractServers%d", i)} }},
    {"components", 10000,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("component%d", i)} }},
    {"purposes", 100,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("purpose%d", i)} }},
    {"people", 100,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("%s%d", utilities.GetRandomName(), i)} }},
    {"physicalServers", 1000,  func(i int) map[string]interface{} {
        cores := rand.Intn(32)
        return map[string]interface{}{"_key": fmt.Sprintf("server%d", i), "cores": cores}
    }},
}

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
    if err != nil {
        panic(err)
    }

    for _, collInfo := range collections {
        if coll, err := db.Collection(nil, collInfo.Name); err == nil {
            coll.Remove(nil)
        }

        ctx := context.Background()
        options := &driver.CreateCollectionOptions{}
        _, err := db.CreateCollection(ctx, collInfo.Name, options)
        if err != nil {
            panic(err)
        }

        coll, err := db.Collection(context.Background(), collInfo.Name)
        if err != nil {
            panic(err)
        }

        for i := 0; i < collInfo.Size; i++ {
            doc := collInfo.DocGenFn(i)
            _, err := coll.CreateDocument(context.Background(), doc)
	    if err != nil {
	        panic(err)
	    }
        }


	collection, err := db.Collection(ctx, collInfo.Name)
	if err != nil {
	    panic(err)
	}
	count, err := collection.Count(ctx)
	if err != nil {
	    panic(err)
	}

	fmt.Printf("The number of documents in %s is %d\n", collInfo.Name, count)
    }

 }
