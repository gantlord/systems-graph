package main

import (
    "context"
    "fmt"

    driver "github.com/arangodb/go-driver"
    http "github.com/arangodb/go-driver/http"
)

func checkGraph(db driver.Database) {
    components := getComponents(db)
    edgeCollName := "edges"

    for _, component := range components {
        query := fmt.Sprintf("FOR v, e IN OUTBOUND '%s' %s RETURN v._id", component, edgeCollName)
        cursor, err := db.Query(context.Background(), query, nil)
        if err != nil {
            panic(err)
        }
        defer cursor.Close()

        hasOutgoingEdge := false
        for {
            var neighbor string
            _, err := cursor.ReadDocument(context.Background(), &neighbor)
            if driver.IsNoMoreDocuments(err) {
                break
            } else if err != nil {
                panic(err)
            }
            if neighbor == fmt.Sprintf("abstractServers%d", 0) {
                hasOutgoingEdge = true
                break
            }
            for _, c := range components {
                if neighbor == c {
                    hasOutgoingEdge = true
                    break
                }
            }
            if hasOutgoingEdge {
                break
            }
        }
        if !hasOutgoingEdge {
            fmt.Printf("Error: Component %s does not have an outgoing edge to another component or an abstract server\n", component)
        }
    }
}

func main() {
    conn := createConnection()
    client := createClient(conn)
    db := getDB(client)

    checkGraph(db)
}

