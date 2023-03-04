package main

import (
	"fmt"
	"math/rand"
	"systems-graph/arango_utils"

)


func main() {
	rand.New(rand.NewSource(0))

	conn := arango_utils.CreateConnection()
	client := arango_utils.CreateClient(conn)
	db := arango_utils.GetDB(client)

	for _, collInfo := range arango_utils.Collections {
		ctx := arango_utils.CreateCollectionFromInfo(db, collInfo)
		count := arango_utils.GetDocumentCount(db, ctx, collInfo)
		fmt.Printf("The number of documents in %s is %d\n", collInfo.Name, count)
	}

	edgeColl := arango_utils.CreateEdgeCollection(db, "edges")
	components := arango_utils.GetComponents(db)

	for i := range components {
		for {
			if rand.Intn(100) < arango_utils.ConnectionPct {
				j := rand.Intn(len(components))
				if i == j {
					continue
				}
				arango_utils.AttemptEdgeCreation(db, components, i, j, edgeColl)
			} else {
				break
			}
		}

	}

	subgraphCount := arango_utils.GetSubgraphCount(components, db)
	fmt.Printf("Number of subgraphs: %d\n", subgraphCount)
        arango_utils.CheckGraph(db)
}

