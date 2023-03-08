package arango_utils

import (
	"context"
	"fmt"
	"math/rand"
	driver "github.com/arangodb/go-driver"
	http "github.com/arangodb/go-driver/http"
)

const small = 10
const medium = 10  * small
const large = 10 * medium
const ConnectionPct = 50

var AuditsAllSucceeded = true

type CollectionInfo struct {
	Name     string
	Size     int
	DocGenFn func(i int) map[string]interface{}
}

var Collections = []CollectionInfo{
	{"binaries", medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("binary%d", i)} }},
	{"firewallRules", small, func(i int) map[string]interface{} { 
		instances := rand.Intn(small)
		return map[string]interface{}{"_key": fmt.Sprintf("fireWallRule%d", i), "instances": instances} 
	}},
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



func CreateEdge(db driver.Database, from string, to string, edgeColl driver.Collection, panic_if_cycle bool) {
	if !createsCycle(db, "edges", from, to, panic_if_cycle) {

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
	query := fmt.Sprintf("FOR vertex IN %s RETURN vertex._id", collection)
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

func createsCycle(db driver.Database, edgeCollName string, fromVertex string, toVertex string, panic_if_cycle bool) bool {
	//TODO check that this method actually works - looks reasonably efficient - but is it right? needs a test
	//TODO almost certain to be broken, no max depth specified
	visitedVertices := make(map[string]bool)
	visitedVertices[fromVertex] = true
	query := fmt.Sprintf(`
		FOR v, e IN ANY '%s' %s
			FILTER e._to == '%s'
			RETURN v._id
	`, fromVertex, edgeCollName, toVertex)
	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	cycle_exists := cursor.Count() > 0
	if cycle_exists {
		fmt.Printf("cycle in %s to %s\n", fromVertex, toVertex)
		panic("cycle")
	}

	return cycle_exists
}

func GetRandomName() string {
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]
	return fmt.Sprintf("%s%s", firstName, lastName)
}

func CheckVertexHasOutgoingEdgeToCollection(db driver.Database, vertex string, edgeCollName string, collection []string) bool {
	//TODO check this is sane - might be the only working one
        query := fmt.Sprintf("FOR v, e IN OUTBOUND '%s' %s RETURN v._id", vertex, edgeCollName)
        cursor, err := db.Query(context.Background(), query, nil)
        if err != nil {
            panic(err)
        }
        defer cursor.Close()

        hasOutgoingEdge := false
        for {
            var neighbour string
            _, err := cursor.ReadDocument(context.Background(), &neighbour)
            if driver.IsNoMoreDocuments(err) {
                break
            } else if err != nil {
                panic(err)
            }
            for _, c := range collection {
                if neighbour == c {
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
	subgraphCount := GetSubgraphCount(db, collInfo.Name)

	if subgraphCount == 1 {
            fmt.Printf("SUCCESS: collection %s is fully connected for %d vertices\n", collInfo.Name, collInfo.Size)
	} else {
	    fmt.Printf("FAILURE: collection %s has %d subgraphs for %d vertices\n", collInfo.Name, subgraphCount, collInfo.Size)
	    AuditsAllSucceeded = false
	}
}



func subgraphHasConnectionToCollection(db driver.Database, edgeCollName string, vertex string, visitedVertices map[string]bool, targetCollectionName string) bool {
	visitedVertices[vertex] = true
        
	targetCollectionIDs := GetCollectionIDsAsString(db, targetCollectionName)
        targetCollectionConnectionExists := CheckVertexHasOutgoingEdgeToCollection(db, vertex, edgeCollName, targetCollectionIDs)

	//TODO may not work without proper max depth setting
	query := fmt.Sprintf("FOR v, e IN ANY '%s' %s RETURN v._id", vertex, edgeCollName)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		panic(err)
	}
	defer cursor.Close()
	for {
		var neighbour string
		_, err := cursor.ReadDocument(context.Background(), &neighbour)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			panic(err)
		}
		if !visitedVertices[neighbour] {
                    result := subgraphHasConnectionToCollection(db, edgeCollName, neighbour, visitedVertices, targetCollectionName)
	            targetCollectionConnectionExists = result || targetCollectionConnectionExists
		}
	}

	return targetCollectionConnectionExists

}

func GetSubgraphCount(db driver.Database, collectionName string) int {
	//TODO this query hangs if run under certain circumstances, not yet determined
	query := fmt.Sprintf("let finalArray=(for x in %s let subResult=(for v in 0..10 any x edges options {\"uniqueVertices\":\"path\"} filter is_same_collection(\"%s\", v._id) collect keys=v._key into found return distinct keys) return distinct subResult) return length(finalArray)", collectionName)
	return queryIntResult(db, query)
}

func GetSubgraphConnectionsToTargetCollectionCount(db driver.Database, sourceCollectionName string, targetCollectionName string) int {
   //TODO may be broken, no max depth 
	query := fmt.Sprintf("let pArray=(for x in %s let countParts=(for v in outbound x edges filter is_same_collection(\"%s\", v._id) collect with count into cCount let part=cCount==0? 0:1 return part) return first(countParts)) return sum(flatten(pArray))", sourceCollectionName, targetCollectionName)
	result := queryIntResult(db, query)
	
	return result
}

func AuditCollectionSubgraphsConnectToCollection(db driver.Database, sourceCollectionName string, targetCollectionName string) {
    
    subgraphConnectionsCount := GetSubgraphConnectionsToTargetCollectionCount(db, sourceCollectionName, targetCollectionName)
    subgraphCount := GetSubgraphCount(db, sourceCollectionName)

    if subgraphConnectionsCount!=subgraphCount {
        fmt.Printf("FAILURE: %d/%d subgraphs in source collection %s have an outgoing edge to collection %s\n", subgraphConnectionsCount, subgraphCount, sourceCollectionName, targetCollectionName) 
	AuditsAllSucceeded = false
    } else {
        fmt.Printf("SUCCESS: %d/%d subgraphs in source collection %s have an outgoing edge to target collection %s\n", subgraphConnectionsCount, subgraphCount, sourceCollectionName, targetCollectionName) 
    }
}

func queryIntResult(db driver.Database, queryString string) int {
    cursor, err := db.Query(context.Background(), queryString, nil)
    if err != nil {
        panic(err)
    }
    defer cursor.Close()
    var result int
    _, err = cursor.ReadDocument(context.Background(), &result)
    if driver.IsNoMoreDocuments(err) {
    } else if err != nil {
	panic(err)
    }
    return result
}

func AuditAllVerticesConnectToCollection(db driver.Database, sourceCollectionName string, targetCollectionName string, sourceCollectionLength int) {

    query := fmt.Sprintf("let pArray=(for x in %s let countParts=(for v in outbound x edges filter is_same_collection(\"%s\", v._id) collect with count into cCount let part=cCount==0? 0:1 return part) return first(countParts)) return sum(flatten(pArray))", sourceCollectionName, targetCollectionName)
    
//TODO may be broken, no max depth

    result := queryIntResult(db, query)
   
    resString := "FAILURE"
    if result==sourceCollectionLength{
        resString = "SUCCESS"
    } else {
	AuditsAllSucceeded = false
    }

    fmt.Printf("%s: %d/%d vertices in source collection %s have an outgoing edge to collection %s\n", resString, result, sourceCollectionLength, sourceCollectionName, targetCollectionName) 
}


func AuditComponentsConnectToEitherCollection(db driver.Database, sourceCollectionName string, targetCollectionName1 string, targetCollectionName2 string, sourceCollectionLength int) {

//TODO no max depth

    query := fmt.Sprintf("let pArray=(for x in %s let countParts=(for v in outbound x edges filter is_same_collection(\"%s\", v._id) || is_same_collection(\"%s\", v._id) collect with count into cCount let part=cCount==0? 0:1 return part) return first(countParts)) return sum(flatten(pArray))", sourceCollectionName, targetCollectionName1, targetCollectionName2)
    result := queryIntResult(db, query)

    resString := "FAILURE"
    if result==sourceCollectionLength{
        resString = "SUCCESS"
    } else {
	AuditsAllSucceeded = false
    }

    fmt.Printf("%s: %d/%d vertices in source collection %s have an outgoing edge to collection %s or collection %s\n", resString, result, sourceCollectionLength, sourceCollectionName, targetCollectionName1, targetCollectionName2) 

}

