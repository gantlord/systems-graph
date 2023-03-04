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

	for _, collInfo := range arango_utils.Collections {
		arango_utils.CreateCollectionFromInfo(db, collInfo)
	}

	edgeColl := arango_utils.CreateEdgeCollection(db, "edges")
	components := arango_utils.GetCollectionIDsAsString(db, "components")
	pods := arango_utils.GetCollectionIDsAsString(db, "pods")
	binaries := arango_utils.GetCollectionIDsAsString(db, "binaries")
	firewallRules := arango_utils.GetCollectionIDsAsString(db, "firewallRules")
	purposes := arango_utils.GetCollectionIDsAsString(db, "purposes")
	people := arango_utils.GetCollectionIDsAsString(db, "people")
	nodes := arango_utils.GetCollectionIDsAsString(db, "nodes")

	for i := range components {
		arango_utils.CreateEdge(db, components[i], binaries[rand.Intn(len(binaries))], edgeColl)
		for {
			if rand.Intn(100) < arango_utils.ConnectionPct {
				j := rand.Intn(len(components))
				if i == j {
					continue
				}
				arango_utils.CreateEdge(db, components[i], components[j], edgeColl)
			} else {
				arango_utils.CreateEdge(db, components[i], pods[rand.Intn(len(pods))], edgeColl)
				break
			}
		}

	}

	for i := range firewallRules {
		arango_utils.CreateEdge(db, components[rand.Intn(len(components))], firewallRules[i], edgeColl)
	}

	for i := range purposes {
		arango_utils.CreateEdge(db, components[rand.Intn(len(components))], purposes[i], edgeColl)
	}

	for i := range people {
		arango_utils.CreateEdge(db, binaries[i], people[i], edgeColl)
	}

	for i := range nodes {
		arango_utils.CreateEdge(db, pods[i], nodes[i], edgeColl)
	}

	for _, collInfo := range arango_utils.Collections {
	    arango_utils.AuditCollectionIsFullyConnected(collInfo, db)
	}

	arango_utils.AuditComponentsConnectToComponentOrPod(db)
	arango_utils.AuditCollectionSubgraphsConnectToCollection(db, "components", "pods")
	arango_utils.AuditCollectionSubgraphsConnectToCollection(db, "components", "purposes")

	if (arango_utils.AuditsAllSucceeded){
	    fmt.Println("\nAll audits completed successfully!")
	    os.Exit(0)
	} else {
	    fmt.Println("\nAudits failed")
	    os.Exit(1)
 	}

}

