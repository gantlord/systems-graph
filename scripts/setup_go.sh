sudo yum -y install go;
cd go; 
go mod init test
go get github.com/arangodb/go-driver
go build
./test
