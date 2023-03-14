package main

import (
	"math/rand"
	"systems-graph/sg_utils"
	"fmt"
	"os"
)




func main() {
	rand.New(rand.NewSource(0))

	db := sg_utils.GetDB()
	sg_utils.DeleteDB(db)

	IDsMap := make(map[string][]string)

	fmt.Println("Creating Vertices...")
	for _, labelInfo := range sg_utils.Labels {
		IDsMap[labelInfo.Name] = sg_utils.CreateVerticesFromInfo(db, labelInfo)
	}

	components := IDsMap["components"]
	pods := IDsMap["pods"]
	binaries := IDsMap["binaries"]
	firewallRules := IDsMap["firewallRules"]
	purposes := IDsMap["purposes"]
	people := IDsMap["people"]
	nodes := IDsMap["nodes"]

	/*TODO Audits:
	6. firewallRules don't have too many instances
	7. components don't consume too many cores on node */

	fmt.Println("Creating Edges...")

	for i := range components {
		sg_utils.CreateEdge(db, "components", "purposes", components[i], purposes[rand.Intn(len(purposes))], "HAS_PURPOSE")
		sg_utils.CreateEdge(db, "components", "binaries", components[i], binaries[rand.Intn(len(binaries))], "INSTANCE_OF")
		sg_utils.CreateEdge(db, "components", "pods", components[i], pods[rand.Intn(len(pods))], "COMPONENT_MAPPED_TO")
	}

	for i := range firewallRules {
		sg_utils.CreateEdge(db, "components", "firewallRules", components[rand.Intn(len(components))], firewallRules[i], "NEEDS_FW_RULE")
	}

	for i := range binaries {
		sg_utils.CreateEdge(db, "binaries", "people", binaries[i], people[rand.Intn(len(binaries))], "MAINTAINED_BY")
	}

	for i := range pods {
		sg_utils.CreateEdge(db, "pods", "nodes", pods[i], nodes[rand.Intn(len(pods))], "POD_MAPPED_TO")
	}

	fmt.Println("Auditing Relationships...")
	sg_utils.AuditAllVerticesConnectToLabel(db, "components", "pods", "COMPONENT_MAPPED_TO", len(components))
	sg_utils.AuditAllVerticesConnectToLabel(db, "components", "purposes", "HAS_PURPOSE", len(components))
	sg_utils.AuditAllVerticesConnectToLabel(db, "components", "binaries", "INSTANCE_OF", len(components))
	sg_utils.AuditAllVerticesConnectFromLabel(db, "firewallRules", "components", "NEEDS_FW_RULE", len(firewallRules))
	sg_utils.AuditAllVerticesConnectToLabel(db, "binaries", "people", "MAINTAINED_BY", len(binaries))
	sg_utils.AuditAllVerticesConnectToLabel(db, "pods", "nodes", "POD_MAPPED_TO", len(pods))

	if (sg_utils.AuditsAllSucceeded){
	    fmt.Println("\nAll audits completed successfully!")
	    os.Exit(0)
	} else {
	    fmt.Println("\nAudits failed")
	    os.Exit(1)
 	}

}
