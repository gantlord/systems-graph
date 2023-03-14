package neo_utils

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func CreateDriver() neo4j.Driver {
	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("username", "password", ""))
	if err != nil {
		panic(err)
	}
	return driver
}

func GetDB(driver neo4j.Driver) neo4j.Session {
	neo4jSession := driver.NewSession(neo4j.SessionConfig{})
	return neo4jSession
}
