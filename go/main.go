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

var small = 10
var medium = 10 * small
var large = 10 * medium

var collections = []CollectionInfo{
    {"binaries", medium,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("binary%d", i)} }},
    {"firewallRules", small,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("rule%d", i)} }},
    {"abstractServers", medium,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("abstractServers%d", i)} }},
    {"components", large,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("component%d", i)} }},
    {"purposes", medium,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("purpose%d", i)} }},
    {"people", medium,  func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("%s%d", utilities.GetRandomName(), i)} }},
    {"physicalServers", medium,  func(i int) map[string]interface{} {
        cores := rand.Intn(32)
        return map[string]interface{}{"_key": fmt.Sprintf("server%d", i), "cores": cores}
    }},
}

var connectionPct = 50

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


    // Randomly connect nodes without creating a cycle
	rand.Seed(0)

	edgeCollName := "edges"

        if edgeColl, err := db.Collection(nil, edgeCollName); err == nil {
            edgeColl.Remove(nil)
        }
	// Create the edge collection if it doesn't exist
	opts := &driver.CreateCollectionOptions{
		Type: driver.CollectionTypeEdge,
	}
	edgeColl, err := db.CreateCollection(context.Background(), edgeCollName, opts)
	if err != nil {
		panic(err)
	}

	// Get all nodes in the collection
	var nodes []string
	query := fmt.Sprintf("FOR node IN %s RETURN node._id", "components")
	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	for {
		var node string
		_, err := cursor.ReadDocument(context.Background(), &node)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			panic(err)
		}
		nodes = append(nodes, node)
	}

	if err != nil {
		panic(err)
	}

	// Randomly connect nodes
	for i := range nodes {
		for {
			if rand.Intn(100) < connectionPct {
				j := rand.Intn(len(nodes)) 
				if i == j {
					continue
				}
				//fmt.Printf("i=%d, j=%d\n", i, j)
				// Check if the connection creates a cycle
				if !createsCycle(db, "edges", nodes[i], nodes[j]) {

					edge := map[string]interface{}{
						"_from": fmt.Sprintf("%s", nodes[i]),
						"_to":   fmt.Sprintf("%s", nodes[j]),
					}
					_, err := edgeColl.CreateDocument(context.Background(), edge)
					//§fmt.Println(edge)
					if err != nil {
						panic(err)
					}
				}
			} else {
				break
			}
		}

	}

	// Count the number of subgraphs in the nodes
	var visitedNodes = make(map[string]bool)
	var subgraphCount int
	for _, node := range nodes {
		if !visitedNodes[node] {
			subgraphCount++
			dfs(db, "edges", node, visitedNodes)
		}
	}
	fmt.Printf("Number of subgraphs: %d\n", subgraphCount)

 }

// dfs performs depth-first search on the graph starting from a given node
func dfs(db driver.Database, edgeCollName string, node string, visitedNodes map[string]bool) {
	visitedNodes[node] = true
	query := fmt.Sprintf("FOR v, e IN ANY '%s' %s RETURN v._id", node, edgeCollName)
	cursor, err := db.Query(context.Background(), query, nil)
	//fmt.Println(query)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	for {
		var neighbor string
		_, err := cursor.ReadDocument(context.Background(), &neighbor)
		//fmt.Println(neighbor)
		//fmt.Println(err)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			panic(err)
		}
		if !visitedNodes[neighbor] {
			dfs(db, edgeCollName, neighbor, visitedNodes)
		}
	}
}

// createsCycle checks if connecting two nodes will create a cycle in the graph
func createsCycle(db driver.Database, edgeCollName string, fromNode string, toNode string) bool {
	visitedNodes := make(map[string]bool)
	visitedNodes[fromNode] = true
	query := fmt.Sprintf(`
		FOR v, e IN ANY '%s' %s
			FILTER e._to == '%s'
			RETURN v._id
	`, fromNode, edgeCollName, toNode)
	cursor, err := db.Query(context.Background(), query, nil)
	//fmt.Println(query)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	if cursor.Count() > 0 {
		return true
	}
	return false
}
