package main

import (
	"math/rand"
	"systems-graph/sg_utils"
	"fmt"
	"os"
)




func main() {
	rand.New(rand.NewSource(0))

    	config := sg_utils.ParseConfig()
	db := sg_utils.GetDB(config)
	sg_utils.DeleteDB(db)

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
		sg_utils.CreateEdge(db, "components", "purposes", components[i], purposes[rand.Intn(len(purposes))], "HAS_PURPOSE", edgeColl, false)
		sg_utils.CreateEdge(db, "components", "binaries", components[i], binaries[rand.Intn(len(binaries))], "INSTANCE_OF", edgeColl, false)
		for {
			if rand.Intn(100) < sg_utils.ConnectionPct {
				j := rand.Intn(len(components))
				if i == j {
					continue
				}
				//sg_utils.CreateEdge(db, "components", "components", components[i], components[j], "DEPENDS_ON", edgeColl, false)
			} else {
				sg_utils.CreateEdge(db, "components", "pods", components[i], pods[rand.Intn(len(pods))], "RESIDES_ON", edgeColl, false)
				break
			}
		}

	}

	for i := range firewallRules {
		sg_utils.CreateEdge(db, "components", "firewallRules", components[rand.Intn(len(components))], firewallRules[i], "NEEDS_FW_RULE", edgeColl, false)
	}

	for i := range people {
		sg_utils.CreateEdge(db, "binaries", "people", binaries[i], people[i], "MAINTAINED_BY", edgeColl, false)
	}

	for i := range nodes {
		sg_utils.CreateEdge(db, "pods", "nodes", pods[i], nodes[i], "MAPPED_TO", edgeColl, false)
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

