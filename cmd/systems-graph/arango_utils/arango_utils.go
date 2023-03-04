package arango_utils

import (
	"context"
	"fmt"
	"math/rand"
	driver "github.com/arangodb/go-driver"
	http "github.com/arangodb/go-driver/http"
)

const small = 10
const medium = 10 * small
const large = 10 * medium
const ConnectionPct = 50

type CollectionInfo struct {
	Name     string
	Size     int
	DocGenFn func(i int) map[string]interface{}
}

var Collections = []CollectionInfo{
	{"binaries", medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("binary%d", i)} }},
	{"firewallRules", small, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("rule%d", i)} }},
	{"abstractServers", medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("abstractServers%d", i)}
	}},
	{"components", large, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("component%d", i)}
	}},
	{"purposes", medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("purpose%d", i)} }},
	{"people", medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("%s%d", GetRandomName(), i)}
	}},
	{"physicalServers", medium, func(i int) map[string]interface{} {
		cores := rand.Intn(32)
		return map[string]interface{}{"_key": fmt.Sprintf("server%d", i), "cores": cores}
	}},
}

var firstNames = []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Julia", "Kevin", "Linda", "Mallory", "Nancy", "Oscar", "Peggy", "Quentin", "Randy", "Sybil", "Trent", "Ursula", "Victor", "Wendy", "Xander", "Yvonne", "Zelda"}
var lastNames = []string{"Smith", "Johnson", "Brown", "Davis", "Wilson", "Kim", "Schmidt", "Petrov", "Rodriguez", "Garcia", "Gonzalez", "Martinez", "Hernandez", "Lopez", "Perez", "Jackson", "Taylor", "Lee", "Nguyen", "Chen", "Wang", "Singh", "Kim", "Gupta", "Kumar"}


func GetSubgraphCount(components []string, db driver.Database) int {
	var visitedNodes = make(map[string]bool)
	var subgraphCount int
	for _, node := range components {
		if !visitedNodes[node] {
			subgraphCount++
			dfs(db, "edges", node, visitedNodes)
		}
	}
	return subgraphCount
}

func AttemptEdgeCreation(db driver.Database, components []string, i int, j int, edgeColl driver.Collection) {
	if !createsCycle(db, "edges", components[i], components[j]) {

		edge := map[string]interface{}{
			"_from": components[i],
			"_to":   components[j],
		}
		_, err := edgeColl.CreateDocument(context.Background(), edge)
		if err != nil {
			panic(err)
		}
	}
}

func GetComponents(db driver.Database) []string {
	var components []string
	query := fmt.Sprintf("FOR node IN %s RETURN node._id", "components")
	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	for {
		var component string
		_, err := cursor.ReadDocument(context.Background(), &component)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			panic(err)
		}
		components = append(components, component)
	}

	if err != nil {
		panic(err)
	}
	return components
}

func CreateEdgeCollection(db driver.Database, edgeCollName string) driver.Collection {
	if edgeColl, err := db.Collection(context.TODO(), edgeCollName); err == nil {
		edgeColl.Remove(context.TODO())
	}
	opts := &driver.CreateCollectionOptions{
		Type: driver.CollectionTypeEdge,
	}
	edgeColl, err := db.CreateCollection(context.Background(), edgeCollName, opts)
	if err != nil {
		panic(err)
	}
	return edgeColl
}

func GetDocumentCount(db driver.Database, ctx context.Context, collInfo CollectionInfo) int64 {
	collection, err := db.Collection(ctx, collInfo.Name)
	if err != nil {
		panic(err)
	}
	count, err := collection.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

func CreateCollectionFromInfo(db driver.Database, collInfo CollectionInfo) context.Context {
	if coll, err := db.Collection(context.Background(), collInfo.Name); err == nil {
		coll.Remove(context.TODO())
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
	return ctx
}

func GetDB(client driver.Client) driver.Database {
	db, err := client.Database(context.TODO(), "_system")
	if err != nil {
		panic(err)
	}
	return db
}

func CreateClient(conn driver.Connection) driver.Client {
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("username", "password"),
	})
	if err != nil {
		panic(err)
	}
	return client
}

func CreateConnection() driver.Connection {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		panic(err)
	}
	return conn
}

func dfs(db driver.Database, edgeCollName string, node string, visitedNodes map[string]bool) {
	visitedNodes[node] = true
	query := fmt.Sprintf("FOR v, e IN ANY '%s' %s RETURN v._id", node, edgeCollName)
	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	for {
		var neighbor string
		_, err := cursor.ReadDocument(context.Background(), &neighbor)
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

func createsCycle(db driver.Database, edgeCollName string, fromNode string, toNode string) bool {
	visitedNodes := make(map[string]bool)
	visitedNodes[fromNode] = true
	query := fmt.Sprintf(`
		FOR v, e IN ANY '%s' %s
			FILTER e._to == '%s'
			RETURN v._id
	`, fromNode, edgeCollName, toNode)
	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	return cursor.Count() > 0
}

func GetRandomName() string {
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]
	return fmt.Sprintf("%s%s", firstName, lastName)
}

func CheckGraph(db driver.Database) {
    components := GetComponents(db)
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
