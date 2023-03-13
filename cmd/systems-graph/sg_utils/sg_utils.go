package sg_utils

import (
	"context"
	"fmt"
	"strings"
	"math/rand"
	"os"
	"systems-graph/neo_utils"
	"systems-graph/arango_utils"
	arango_driver "github.com/arangodb/go-driver"
	neo_driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	//http "github.com/arangodb/go-driver/http"

)

const Small = 1
const Medium = 10  * Small
const Large = 10 * Medium
const ConnectionPct = 50

var AuditsAllSucceeded = true
var DefaultMaxDepth = 20

type DBType int

const (
    Neo4j DBType = iota
    ArangoDB
)

func (dbType DBType) String() string {
    switch dbType {
    case Neo4j:
        return "neo4j"
    case ArangoDB:
        return "arangodb"
    default:
        panic(fmt.Sprintf("Unknown database type: %d", dbType))
    }
}

type Config struct {
    dbType DBType
}


func ParseConfig() Config {
    if len(os.Args) != 2 {
        fmt.Println("Usage: ./program-name [neo4j|arangodb]")
        os.Exit(1)
    }

    var dbType DBType
    switch os.Args[1] {
    case "neo4j":
        dbType = Neo4j
    case "arangodb":
        dbType = ArangoDB
    default:
        fmt.Println("Usage: ./program-name [neo4j|arangodb]")
        os.Exit(1)
    }

    var config Config
    config.dbType = dbType
    	
    fmt.Printf("Selected DB type: %s\n", config.dbType)

    return config
}

type CollectionInfo struct {
	Name     string
	Size     int
	DocGenFn func(i int) map[string]interface{}
}

var Collections = []CollectionInfo{
	{"binaries", Medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("binary%d", i)} }},
	{"firewallRules", Small, func(i int) map[string]interface{} { 
		instances := rand.Intn(Small)
		return map[string]interface{}{"_key": fmt.Sprintf("fireWallRule%d", i), "instances": fmt.Sprintf("%d", instances)} 
	}},
	{"pods", Medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("pod%d", i)}
	}},
	{"components", Large, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("component%d", i)}
	}},
	{"purposes", Medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("purpose%d", i)} }},
	{"nodes", Medium, func(i int) map[string]interface{} {
		cores := rand.Intn(32)
		return map[string]interface{}{"_key": fmt.Sprintf("node%d", i), "cores": fmt.Sprintf("%d", cores)}
	}},
	{"people", Medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("%s%d", GetRandomName(), i)}
	}},
}

func stringFromDocFn(doc map[string]interface{}) string {
    var builder strings.Builder
    builder.WriteString("{ ")

    first := true
    for key, value := range doc {
 	//TODO find neater way to do this
	if !first {
		builder.WriteString(fmt.Sprintf(", "))
	}
	first = false
        builder.WriteString(fmt.Sprintf("%s:\"%s\"", key, value))
    }

    builder.WriteString("}")
    return builder.String()
}

var firstNames = []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Julia", "Kevin", "Linda", "Mallory", "Nancy", "Oscar", "Peggy", "Quentin", "Randy", "Sybil", "Trent", "Ursula", "Victor", "Wendy", "Xander", "Yvonne", "Zelda"}
var lastNames = []string{"Smith", "Johnson", "Brown", "Davis", "Wilson", "Kim", "Schmidt", "Petrov", "Rodriguez", "Garcia", "Gonzalez", "Martinez", "Hernandez", "Lopez", "Perez", "Jackson", "Taylor", "Lee", "Nguyen", "Chen", "Wang", "Singh", "Kim", "Gupta", "Kumar"}

type Connection struct {
	arangoConnection arango_driver.Connection
	neoDriver neo_driver.Driver
	dbType DBType
}

type Client struct {
	connection Connection
	arangoClient arango_driver.Client
	neoDriver neo_driver.Driver
}

type Database struct {
	client Client
	arangoDatabase arango_driver.Database
	neoSession neo_driver.Session
	dbType DBType
}

func CreateConnection(config Config) Connection {
	var connection Connection
	connection.dbType = config.dbType
	if connection.dbType==ArangoDB {
		connection.arangoConnection = arango_utils.CreateConnection()
	} else {
		connection.neoDriver = neo_utils.CreateDriver()
	}
	return connection
}

func CreateClient(conn Connection) Client {
	var client Client
	client.connection = conn
	if client.connection.dbType==ArangoDB {
		client.arangoClient = arango_utils.CreateClient(conn.arangoConnection)
	} else {
		client.neoDriver = conn.neoDriver
	}
	return client
}

func GetDB(config Config) Database {
	connection := CreateConnection(config)
	client := CreateClient(connection)
	var db Database
	db.dbType = connection.dbType
	if connection.dbType==ArangoDB {
		db.arangoDatabase = arango_utils.GetDB(client.arangoClient)
	} else {
		db.neoSession = neo_utils.GetDB(client.neoDriver)
	}
	return db
}

func CreateEdge(db Database, fS string, tS string, from string, to string, edgeLabel string, edgeColl arango_driver.Collection, panic_if_cycle bool) {
	if db.dbType==ArangoDB && !createsCycle(db, "edges", from, to, panic_if_cycle) {

		edge := map[string]interface{}{
			"_from": from,
			"_to":   to,
		}
		_, err := edgeColl.CreateDocument(context.Background(), edge)
		if err != nil {
			panic(err)
		}
	} else {
		query := fmt.Sprintf("match (a:%s {_key:\"%s\"}) match (b:%s {_key:\"%s\"}) merge (a)-[:%s]->(b)", fS, from, tS, to, edgeLabel)
				//TODO apparently Run isn't as proper as ExecuteWrite so should use that down the track
			    result,err := db.neoSession.Run(query, nil)
			    if err != nil {
				    //TODO need autoformatter
				    panic(err)
			    }
			    result.Consume()
		//fmt.Printf("%s\n", query) 	
	}
}

func CreateEdgeCollection(db Database, edgeCollName string) arango_driver.Collection {
	if db.dbType==Neo4j {
		//TODO needs cleaned up for neo4j
		return nil
	}
	if edgeColl, err := db.arangoDatabase.Collection(context.TODO(), edgeCollName); err == nil {
		edgeColl.Remove(context.TODO())
	}
	opts := &arango_driver.CreateCollectionOptions{
		Type: arango_driver.CollectionTypeEdge,
	}
	edgeColl, err := db.arangoDatabase.CreateCollection(context.Background(), edgeCollName, opts)
	if err != nil {
		panic(err)
	}
	return edgeColl
}


func DeleteNeoDB(db Database) {
	cypher := "MATCH (d) DETACH DELETE d"
	_, err := db.neoSession.WriteTransaction(func(tx neo_driver.Transaction) (interface{}, error) {
		return tx.Run(cypher, nil)
	})
	if err != nil {
		panic(err)
	}
}

func DeleteDB(db Database) {
	if db.dbType==Neo4j {
		DeleteNeoDB(db)
	}
}

func CreateCollectionFromInfo(db Database, collInfo CollectionInfo) []string {
	docIDs := make([]string, collInfo.Size)
	var coll arango_driver.Collection
	var err error

	//TODO more efficient deletion could go in deleteDB
	if db.dbType==ArangoDB {
		if coll, err = db.arangoDatabase.Collection(context.Background(), collInfo.Name); err == nil {
			coll.Remove(context.TODO())
		}
		ctx := context.Background()
		options := &arango_driver.CreateCollectionOptions{}
		_, err = db.arangoDatabase.CreateCollection(ctx, collInfo.Name, options)
		if err != nil {
			panic(err)
		}

		coll, err = db.arangoDatabase.Collection(context.Background(), collInfo.Name)
		if err != nil {
			panic(err)
		}

	} 

	for i := 0; i < collInfo.Size; i++ {
		doc := collInfo.DocGenFn(i)
		if db.dbType==ArangoDB {
			docMeta, err := coll.CreateDocument(context.Background(), doc)
			if err != nil {
				panic(err)
			}
			docIDs[i] = string(docMeta.ID)
		} else {
			query := fmt.Sprintf("MERGE (n:%s %s)", collInfo.Name, stringFromDocFn(doc))


			    //fmt.Printf("%s\n", query)

				// TODO apparently Run isn't as proper as ExecuteWrite so should use that down the track
			    result,err := db.neoSession.Run(query, nil)
			    if err != nil {
				    //TODO need autoformatter
				    panic(err)
			    }
			    result.Consume()

			docIDs[i] = fmt.Sprintf("%s", doc["_key"])

		}

	}

	return docIDs
}

func createsCycle(db Database, edgeCollName string, fromVertex string, toVertex string, panic_if_cycle bool) bool {
	//TODO since this probably doesn't even work for arangodb already, just returning false for neo4j
	if db.dbType==Neo4j {
		return false
	}


	//TODO check that this method actually works - looks reasonably efficient - but is it right? needs a test
	//TODO almost certain to be broken, no max depth specified
	visitedVertices := make(map[string]bool)
	visitedVertices[fromVertex] = true
	query := fmt.Sprintf(`
		FOR v, e IN ANY '%s' %s
			FILTER e._to == '%s'
			RETURN v._id
	`, fromVertex, edgeCollName, toVertex)
	cursor, err := db.arangoDatabase.Query(context.Background(), query, nil)
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

func AuditCollectionIsFullyConnected(collInfo CollectionInfo, db Database){
	if db.dbType==Neo4j {
		os.Exit(0)
	}
	subgraphCount := GetSubgraphCount(db, collInfo.Name)

	if subgraphCount == 1 {
            fmt.Printf("SUCCESS: collection %s is fully connected for %d vertices\n", collInfo.Name, collInfo.Size)
	} else {
	    fmt.Printf("FAILURE: collection %s has %d subgraphs for %d vertices\n", collInfo.Name, subgraphCount, collInfo.Size)
	    AuditsAllSucceeded = false
	}
}

func GetSubgraphCount(db Database, collectionName string) int {
	query := fmt.Sprintf("let finalArray=(for x in %s let subResult=(for v in 0..%d any x edges options {\"uniqueVertices\":\"global\", \"order\":\"bfs\", \"vertexCollections\":\"%s\"} collect keys=v._key return distinct keys) return distinct subResult) return length(finalArray)", collectionName, DefaultMaxDepth, collectionName)
	return queryIntResult(db, query)
}

func GetSubgraphConnectionsToTargetCollectionCount(db Database, sourceCollectionName string, targetCollectionName string) int {
   //TODO may be broken, no max depth 
	query := fmt.Sprintf("let pArray=(for x in %s let countParts=(for v in outbound x edges filter is_same_collection(\"%s\", v._id) collect with count into cCount let part=cCount==0? 0:1 return part) return first(countParts)) return sum(flatten(pArray))", sourceCollectionName, targetCollectionName)
	result := queryIntResult(db, query)
	
	return result
}

func AuditCollectionSubgraphsConnectToCollection(db Database, sourceCollectionName string, targetCollectionName string) {
    
    subgraphConnectionsCount := GetSubgraphConnectionsToTargetCollectionCount(db, sourceCollectionName, targetCollectionName)
    subgraphCount := GetSubgraphCount(db, sourceCollectionName)

    if subgraphConnectionsCount!=subgraphCount {
        fmt.Printf("FAILURE: %d/%d subgraphs in source collection %s have an outgoing edge to collection %s\n", subgraphConnectionsCount, subgraphCount, sourceCollectionName, targetCollectionName) 
	AuditsAllSucceeded = false
    } else {
        fmt.Printf("SUCCESS: %d/%d subgraphs in source collection %s have an outgoing edge to target collection %s\n", subgraphConnectionsCount, subgraphCount, sourceCollectionName, targetCollectionName) 
    }
}

func queryIntResult(db Database, queryString string) int {
    cursor, err := db.arangoDatabase.Query(context.Background(), queryString, nil)
    if err != nil {
        panic(err)
    }
    defer cursor.Close()
    var result int
    _, err = cursor.ReadDocument(context.Background(), &result)
    if arango_driver.IsNoMoreDocuments(err) {
    } else if err != nil {
	panic(err)
    }
    return result
}

func AuditAllVerticesConnectToCollection(db Database, sourceCollectionName string, targetCollectionName string, sourceCollectionLength int) {


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


func AuditComponentsConnectToEitherCollection(db Database, sourceCollectionName string, targetCollectionName1 string, targetCollectionName2 string, sourceCollectionLength int) {

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
