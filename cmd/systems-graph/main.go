package main

import (
	"math/rand"
	"systems-graph/sg_utils"
)

func main() {

	sg_utils.Setup()

	db := sg_utils.GetDB()
	sg_utils.DeleteDB(db)

	IDsMap := make(map[string][]string)

 	sg_utils.LogInfo("Creating Vertices...")
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

	sg_utils.LogInfo("Creating Edges...")

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

	sg_utils.LogInfo("Auditing Relationships...")
	sg_utils.AuditAllVerticesConnectToLabel(db, "components", "pods", "COMPONENT_MAPPED_TO", len(components))
	sg_utils.AuditAllVerticesConnectToLabel(db, "components", "purposes", "HAS_PURPOSE", len(components))
	sg_utils.AuditAllVerticesConnectToLabel(db, "components", "binaries", "INSTANCE_OF", len(components))
	sg_utils.AuditAllVerticesConnectFromLabel(db, "firewallRules", "components", "NEEDS_FW_RULE", len(firewallRules))
	sg_utils.AuditAllVerticesConnectToLabel(db, "binaries", "people", "MAINTAINED_BY", len(binaries))
	sg_utils.AuditAllVerticesConnectToLabel(db, "pods", "nodes", "POD_MAPPED_TO", len(pods))

	sg_utils.AuditLimitsRespected(db, "components", "firewallRules", "-[:NEEDS_FW_RULE]->", "instanceLimit") 
	sg_utils.AuditLimitsRespected(db, "components", "nodes", "-[:COMPONENT_MAPPED_TO]->()-[:POD_MAPPED_TO]->", "cores") 

	if sg_utils.AuditsAllSucceeded {
		sg_utils.LogInfo("All audits completed successfully!")
	} else {
		sg_utils.LogError("Audits failed")
	}

}
