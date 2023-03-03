sudo yum -y install go
WD=`pwd`
cd ..
git clone https://github.com/go-delve/delve
cd delve/
go install github.com/go-delve/delve/cmd/dlv
cd $WD/cmd/systems-graph
go mod init systems-graph
go get github.com/arangodb/go-driver
go build
./systems-graph
