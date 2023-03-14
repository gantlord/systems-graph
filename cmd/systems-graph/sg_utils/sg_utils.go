package sg_utils

import (
	"fmt"
	neo_driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"math/rand"
	"strings"
	"systems-graph/neo_utils"
)

const Small = 10
const Medium = 10 * Small
const Large = 10 * Medium
const ConnectionPct = 50

var AuditsAllSucceeded = true
var DefaultMaxDepth = 20

type LabelInfo struct {
	Name     string
	Size     int
	DocGenFn func(i int) map[string]interface{}
}

var Labels = []LabelInfo{
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

        separator := ""
        for key, value := range doc {
                builder.WriteString(fmt.Sprintf("%s%s:\"%s\"", separator, key, value))
                separator = ", "
        }

        builder.WriteString(" }")
        return builder.String()
}


var firstNames = []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Julia", "Kevin", "Linda", "Mallory", "Nancy", "Oscar", "Peggy", "Quentin", "Randy", "Sybil", "Trent", "Ursula", "Victor", "Wendy", "Xander", "Yvonne", "Zelda"}
var lastNames = []string{"Smith", "Johnson", "Brown", "Davis", "Wilson", "Kim", "Schmidt", "Petrov", "Rodriguez", "Garcia", "Gonzalez", "Martinez", "Hernandez", "Lopez", "Perez", "Jackson", "Taylor", "Lee", "Nguyen", "Chen", "Wang", "Singh", "Kim", "Gupta", "Kumar"}

type Database struct {
	neoSession neo_driver.Session
}

func GetDB() Database {
	var db Database
	driver := neo_utils.CreateDriver()
	db.neoSession = neo_utils.GetDB(driver)
	return db
}

func CreateEdge(db Database, fS string, tS string, from string, to string, edgeLabel string) {

	query := fmt.Sprintf("match (a:%s {_key:\"%s\"}) match (b:%s {_key:\"%s\"}) merge (a)-[:%s]->(b)", fS, from, tS, to, edgeLabel)
	//TODO apparently Run isn't as proper as ExecuteWrite so should use that down the track
	result, err := db.neoSession.Run(query, nil)
	if err != nil {
		panic(err)
	}
	result.Consume()
	//fmt.Printf("%s\n", query)
}

func DeleteDB(db Database) {
	cypher := "MATCH (d) DETACH DELETE d"
	_, err := db.neoSession.WriteTransaction(func(tx neo_driver.Transaction) (interface{}, error) {
		return tx.Run(cypher, nil)
	})
	if err != nil {
		panic(err)
	}
}

func CreateVerticesFromInfo(db Database, labelInfo LabelInfo) []string {
	docIDs := make([]string, labelInfo.Size)

	for i := 0; i < labelInfo.Size; i++ {
		doc := labelInfo.DocGenFn(i)
		query := fmt.Sprintf("MERGE (n:%s %s)", labelInfo.Name, stringFromDocFn(doc))

		//fmt.Printf("%s\n", query)

		// TODO apparently Run isn't as proper as ExecuteWrite so should use that down the track
		result, err := db.neoSession.Run(query, nil)
		if err != nil {
			panic(err)
		}
		result.Consume()

		docIDs[i] = fmt.Sprintf("%s", doc["_key"])
	}

	return docIDs
}

func GetRandomName() string {
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]
	return fmt.Sprintf("%s%s", firstName, lastName)
}

func queryIntResult(db Database, queryString string, resultString string) int {
	//TODO apparently Run isn't as proper as ExecuteWrite so should use that down the track
	result, err := db.neoSession.Run(queryString, nil)
	if err != nil {
		panic(err)
	}
	record, err := result.Single()
	if err != nil {
		panic(err)
	}
	count, exists := record.Get(resultString)
	if !exists {
		panic(fmt.Sprintf("%s doesn't exist", resultString))
	}
	return int(count.(int64))
}

func AuditAllVerticesConnectToLabel(db Database, sourceLabelName string, targetLabelName string, relationshipName string, sourceLabelLength int) {

	query := fmt.Sprintf("match (x:%s)-[:%s]->(y:%s) return count(x)", sourceLabelName, relationshipName, targetLabelName)

	result := queryIntResult(db, query, "count(x)")

	resString := "FAILURE"
	if result == sourceLabelLength {
		resString = "SUCCESS"
	} else {
		AuditsAllSucceeded = false
	}

	fmt.Printf("%s: %d/%d vertices in source label %s have an outgoing edge to label %s\n", resString, result, sourceLabelLength, sourceLabelName, targetLabelName)
}
func AuditAllVerticesConnectFromLabel(db Database, targetLabelName string, sourceLabelName string, relationshipName string, targetLabelLength int) {

	query := fmt.Sprintf("match (x:%s)<-[:%s]-(y:%s) return count(x)", targetLabelName, relationshipName, sourceLabelName)

	result := queryIntResult(db, query, "count(x)")

	resString := "FAILURE"
	if result == targetLabelLength {
		resString = "SUCCESS"
	} else {
		AuditsAllSucceeded = false
	}

	fmt.Printf("%s: %d/%d vertices in target label %s have an incoming edge from label %s\n", resString, result, targetLabelLength, targetLabelName, sourceLabelName)
}
