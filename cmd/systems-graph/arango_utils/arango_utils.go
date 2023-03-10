package arango_utils

import (
	"context"
	driver "github.com/arangodb/go-driver"
	http "github.com/arangodb/go-driver/http"
)

func GetDB(client driver.Client) driver.Database {
	db, err := client.Database(context.TODO(), "_system")
	if err != nil {
		panic(err)
	}
	return db
}

func CreateClient(conn driver.Connection) driver.Client {
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("username", "password"),
	})
	if err != nil {
		panic(err)
	}
	return client
}

func CreateConnection() driver.Connection {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		panic(err)
	}
	return conn
}

