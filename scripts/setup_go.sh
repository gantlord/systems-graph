sudo yum -y install go
cd go ; go get github.com/arangodb/go-driver ; cd -
cd go ; go mod init test ; go build ; ./test ; cd -
