package main

import (
	"math/rand"
	"systems-graph/arango_utils"

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

	for i := range components {
		for {
			if rand.Intn(100) < arango_utils.ConnectionPct {
				j := rand.Intn(len(components))
				if i == j {
					continue
				}
				arango_utils.AttemptEdgeCreation(db, components[i], components[j], edgeColl)
			} else {
				arango_utils.AttemptEdgeCreation(db, components[i], pods[rand.Intn(len(pods))], edgeColl)
				break
			}
		}

	}

	for _, collInfo := range arango_utils.Collections {
	    arango_utils.AuditCollectionIsFullyConnected(collInfo, db)
	}

	arango_utils.AuditComponentsConnectToComponentOrPod(db)

}

