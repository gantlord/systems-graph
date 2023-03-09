package main

import (
	"math/rand"
	"systems-graph/arango_utils"
	"fmt"
	"os"
)


func main() {
	rand.New(rand.NewSource(0))

	conn := arango_utils.CreateConnection()
	client := arango_utils.CreateClient(conn)
	db := arango_utils.GetDB(client)

	IDsMap := make(map[string][]string)

	for _, collInfo := range arango_utils.Collections {
		IDsMap[collInfo.Name] = arango_utils.CreateCollectionFromInfo(db, collInfo)
	}

	edgeColl := arango_utils.CreateEdgeCollection(db, "edges")
	components := IDsMap["components"]
	pods := IDsMap["pods"]
	binaries := IDsMap["binaries"]
	firewallRules := IDsMap["firewallRules"]
	purposes := IDsMap["purposes"]
	people := IDsMap["people"]
	nodes := IDsMap["nodes"]

	for i := range components {
		//TODO make the purpose tagging more realistic / sparse, invert edge direction
		arango_utils.CreateEdge(db, components[i], purposes[rand.Intn(len(purposes))], edgeColl, false)
		arango_utils.CreateEdge(db, components[i], binaries[rand.Intn(len(binaries))], edgeColl, false)
		for {
			if rand.Intn(100) < arango_utils.ConnectionPct {
				j := rand.Intn(len(components))
				if i == j {
					continue
				}
				arango_utils.CreateEdge(db, components[i], components[j], edgeColl, false)
			} else {
				arango_utils.CreateEdge(db, components[i], pods[rand.Intn(len(pods))], edgeColl, false)
				break
			}
		}

	}

	for i := range firewallRules {
		arango_utils.CreateEdge(db, components[rand.Intn(len(components))], firewallRules[i], edgeColl, false)
	}

	for i := range people {
		arango_utils.CreateEdge(db, binaries[i], people[i], edgeColl, false)
	}

	for i := range nodes {
		arango_utils.CreateEdge(db, pods[i], nodes[i], edgeColl, false)
	}

	for _, collInfo := range arango_utils.Collections {
	    arango_utils.AuditCollectionIsFullyConnected(collInfo, db)
	}

	arango_utils.AuditComponentsConnectToEitherCollection(db, "components", "components", "pods", len(components))
	arango_utils.AuditAllVerticesConnectToCollection(db, "components", "pods", len(components))
	arango_utils.AuditCollectionSubgraphsConnectToCollection(db, "components", "purposes")

	if (arango_utils.AuditsAllSucceeded){
	    fmt.Println("\nAll audits completed successfully!")
	    os.Exit(0)
	} else {
	    fmt.Println("\nAudits failed")
	    os.Exit(1)
 	}

}

