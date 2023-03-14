sudo yum -y install go
WD=`pwd`
cd ..
git clone https://github.com/go-delve/delve
cd delve/
go install github.com/go-delve/delve/cmd/dlv
cd $WD/cmd/systems-graph
go mod init systems-graph
go get github.com/neo4j/neo4j-go-driver/v5

go build
./systems-graph
