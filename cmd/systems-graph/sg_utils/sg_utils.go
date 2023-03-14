package sg_utils

import (
	"fmt"
	"io"
	"log"
	neo_driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"math/rand"
	"os"
	"strings"
	"systems-graph/neo_utils"
	"time"
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
		instanceLimit := rand.Intn(Small) + 1
		return map[string]interface{}{"_key": fmt.Sprintf("fireWallRule%d", i), "instanceLimit": fmt.Sprintf("%d", instanceLimit)}
	}},
	{"pods", Medium, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("pod%d", i)}
	}},
	{"components", Large, func(i int) map[string]interface{} {
		return map[string]interface{}{"_key": fmt.Sprintf("component%d", i)}
	}},
	{"purposes", Medium, func(i int) map[string]interface{} { return map[string]interface{}{"_key": fmt.Sprintf("purpose%d", i)} }},
	{"nodes", Medium, func(i int) map[string]interface{} {
		cores := rand.Intn(32) + 26 //TODO this will break at some future point
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
		log.Fatal(err)
	}
	result.Consume()
}

func DeleteDB(db Database) {
	cypher := "MATCH (d) DETACH DELETE d"
	_, err := db.neoSession.WriteTransaction(func(tx neo_driver.Transaction) (interface{}, error) {
		return tx.Run(cypher, nil)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func CreateVerticesFromInfo(db Database, labelInfo LabelInfo) []string {
	docIDs := make([]string, labelInfo.Size)

	for i := 0; i < labelInfo.Size; i++ {
		doc := labelInfo.DocGenFn(i)
		query := fmt.Sprintf("MERGE (n:%s %s)", labelInfo.Name, stringFromDocFn(doc))

		// TODO apparently Run isn't as proper as ExecuteWrite so should use that down the track
		result, err := db.neoSession.Run(query, nil)
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
	}
	record, err := result.Single()
	if err != nil {
		log.Fatal(err)
	}
	count, exists := record.Get(resultString)
	if !exists {
		log.Fatal(fmt.Sprintf("%s doesn't exist", resultString))
	}
	return int(count.(int64))
}

func AuditLimitsRespected(db Database, sourceLabel string, targetLabel string, relationship string, limit string){

	query := fmt.Sprintf(
		"match (s:%s)%s(d:%s) with d, count(s) as instanceCount, toInteger(d.%s) as limit where instanceCount > limit return count(d)",
		sourceLabel, relationship, targetLabel, limit)

	result := queryIntResult(db, query, "count(d)")
	resString := "FAILURE"

	if result == 0 {
		resString = "SUCCESS"
	} else {
		AuditsAllSucceeded = false
	}

	LogInfo(
		fmt.Sprintf(
			"%s: %d vertices in source label %s have exceeded %s limit %s in target label %s", 
			resString, result, sourceLabel, relationship, limit, targetLabel))

}

func AuditAllVerticesConnectToLabel(db Database, sourceLabel string, targetLabel string, relationship string, sourceLabelLength int) {

	query := fmt.Sprintf(
		"match (x:%s)-[:%s]->(y:%s) return count(x)", 
		sourceLabel, relationship, targetLabel)

	result := queryIntResult(db, query, "count(x)")

	resString := "FAILURE"
	if result == sourceLabelLength {
		resString = "SUCCESS"
	} else {
		AuditsAllSucceeded = false
	}

	LogInfo(
		fmt.Sprintf(
			"%s: %d/%d vertices in source label %s have an outgoing edge to label %s", 
			resString, result, sourceLabelLength, sourceLabel, targetLabel))
}
func AuditAllVerticesConnectFromLabel(db Database, targetLabel string, sourceLabel string, relationship string, targetLabelLength int) {

	query := fmt.Sprintf("match (x:%s)<-[:%s]-(y:%s) return count(x)", targetLabel, relationship, sourceLabel)

	result := queryIntResult(db, query, "count(x)")

	resString := "FAILURE"
	if result == targetLabelLength {
		resString = "SUCCESS"
	} else {
		AuditsAllSucceeded = false
	}

	LogInfo(fmt.Sprintf("%s: %d/%d vertices in target label %s have an incoming edge from label %s", resString, result, targetLabelLength, targetLabel, sourceLabel))
}

func LogInfo(message string){
	log.Printf("INFO: %s\n", message)
}

func LogWarning(message string){
	log.Printf("WARNING: %s\n", message)
}

func LogError(message string){
	log.Printf("ERROR: %s\n", message)
}

func Setup(){
	log.SetFlags(log.Ltime)
    	now := time.Now()

    	timestamp := now.Format("2006-01-02_15-04-05")
	filename := "log/"+timestamp + "-systems-graph.log"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    	if err != nil {
        	log.Fatal(err)
    	
	}
  	log.SetOutput(io.MultiWriter(os.Stdout, file))


	seed := int64(0) 
	rand.New(rand.NewSource(seed))
	LogInfo(fmt.Sprintf("prng seed is %d", seed))

}
