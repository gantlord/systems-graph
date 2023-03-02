sudo yum -y install go
cd cmd/systems-graph
go mod init systems-graph
go get github.com/arangodb/go-driver
go build
./systems-graph
