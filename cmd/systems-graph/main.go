package main

import (
	"math/rand"
	"systems-graph/sg_utils"
	"fmt"
	"os"
)


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
    DBType DBType
}


func parseConfig() *Config {
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

    config := &Config{
        DBType: dbType,
    }

    return config
}


func main() {
    config := parseConfig()

    fmt.Printf("Selected DB type: %s\n", config.DBType)


	rand.New(rand.NewSource(0))

	conn := sg_utils.CreateConnection()
	client := sg_utils.CreateClient(conn)
	db := sg_utils.GetDB(client)

	IDsMap := make(map[string][]string)

	for _, collInfo := range sg_utils.Collections {
		IDsMap[collInfo.Name] = sg_utils.CreateCollectionFromInfo(db, collInfo)
	}

	edgeColl := sg_utils.CreateEdgeCollection(db, "edges")
	components := IDsMap["components"]
	pods := IDsMap["pods"]
	binaries := IDsMap["binaries"]
	firewallRules := IDsMap["firewallRules"]
	purposes := IDsMap["purposes"]
	people := IDsMap["people"]
	nodes := IDsMap["nodes"]

	for i := range components {
		//TODO make the purpose tagging more realistic / sparse, invert edge direction
		sg_utils.CreateEdge(db, components[i], purposes[rand.Intn(len(purposes))], edgeColl, false)
		sg_utils.CreateEdge(db, components[i], binaries[rand.Intn(len(binaries))], edgeColl, false)
		for {
			if rand.Intn(100) < sg_utils.ConnectionPct {
				j := rand.Intn(len(components))
				if i == j {
					continue
				}
				sg_utils.CreateEdge(db, components[i], components[j], edgeColl, false)
			} else {
				sg_utils.CreateEdge(db, components[i], pods[rand.Intn(len(pods))], edgeColl, false)
				break
			}
		}

	}

	for i := range firewallRules {
		sg_utils.CreateEdge(db, components[rand.Intn(len(components))], firewallRules[i], edgeColl, false)
	}

	for i := range people {
		sg_utils.CreateEdge(db, binaries[i], people[i], edgeColl, false)
	}

	for i := range nodes {
		sg_utils.CreateEdge(db, pods[i], nodes[i], edgeColl, false)
	}

	for _, collInfo := range sg_utils.Collections {
	    sg_utils.AuditCollectionIsFullyConnected(collInfo, db)
	}

	sg_utils.AuditComponentsConnectToEitherCollection(db, "components", "components", "pods", len(components))
	sg_utils.AuditAllVerticesConnectToCollection(db, "components", "pods", len(components))
	sg_utils.AuditCollectionSubgraphsConnectToCollection(db, "components", "purposes")

	if (sg_utils.AuditsAllSucceeded){
	    fmt.Println("\nAll audits completed successfully!")
	    os.Exit(0)
	} else {
	    fmt.Println("\nAudits failed")
	    os.Exit(1)
 	}

}

