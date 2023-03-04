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
	{"firewallRules", small, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("fireWallRule%d", i)} }},
	{"pods", medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("pod%d", i)}
	}},
	{"components", large, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("component%d", i)}
	}},
	{"purposes", medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("purpose%d", i)} }},
	{"people", medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("%s%d", GetRandomName(), i)}
	}},
	{"nodes", medium, func(i int) map[string]interface{} {
		cores := rand.Intn(32)
		return map[string]interface{}{"_key": fmt.Sprintf("node%d", i), "cores": cores}
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
			depthFirstSearch(db, "edges", node, visitedNodes)
		}
	}
	return subgraphCount
}

func CreateEdge(db driver.Database, from string, to string, edgeColl driver.Collection) {
	if !createsCycle(db, "edges", from, to) {

		edge := map[string]interface{}{
			"_from": from,
			"_to":   to,
		}
		_, err := edgeColl.CreateDocument(context.Background(), edge)
		if err != nil {
			panic(err)
		}
	}
}

func GetCollectionIDsAsString(db driver.Database, collection string) []string {
	var collectionIDs []string
	query := fmt.Sprintf("FOR node IN %s RETURN node._id", collection)
	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	for {
		var collectionID string
		_, err := cursor.ReadDocument(context.Background(), &collectionID)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			panic(err)
		}
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err != nil {
		panic(err)
	}
	return collectionIDs
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

func depthFirstSearch(db driver.Database, edgeCollName string, node string, visitedNodes map[string]bool) {
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
			depthFirstSearch(db, edgeCollName, neighbor, visitedNodes)
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

func CheckVertexHasOutgoingEdgeToCollection(db driver.Database, vertex string, edgeCollName string, collection []string) bool {
        query := fmt.Sprintf("FOR v, e IN OUTBOUND '%s' %s RETURN v._id", vertex, edgeCollName)
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
            for _, c := range collection {
                if neighbor == c {
                    hasOutgoingEdge = true
                    break
                }
            }
            if hasOutgoingEdge {
                break
            }
        }
        return hasOutgoingEdge
}

func AuditCollectionIsFullyConnected(collInfo CollectionInfo, db driver.Database){
	collectionIDs := GetCollectionIDsAsString(db, collInfo.Name)
	subgraphCount := GetSubgraphCount(collectionIDs, db)

	if subgraphCount == 1 {
            fmt.Printf("SUCCESS: collection %s is fully connected for %d vertices\n", collInfo.Name, collInfo.Size)
	} else {
	    fmt.Printf("FAILURE: collection %s has %d subgraphs for %d vertices\n", collInfo.Name, subgraphCount, collInfo.Size)
	}
}

func AuditComponentsConnectToComponentOrPod(db driver.Database) {
    components := GetCollectionIDsAsString(db, "components")
    pods := GetCollectionIDsAsString(db, "pods")
    edgeCollName := "edges"
    numFailures := 0	

    for _, component := range components {
	hasOutgoingEdgeComponents := CheckVertexHasOutgoingEdgeToCollection(db, component, edgeCollName, components)
	hasOutgoingEdgePods := CheckVertexHasOutgoingEdgeToCollection(db, component, edgeCollName, pods)
        if !hasOutgoingEdgeComponents && !hasOutgoingEdgePods{
	    numFailures++	
        }
    }
    if numFailures!=0 {
        fmt.Printf("FAILURE: %d component(s) do(es) not have an outgoing edge to another component or a pod\n", numFailures) 
    } else {
        fmt.Println("SUCCESS: all components have an outgoing edge to another component or a pod") 
    }
}

