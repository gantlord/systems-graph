sudo yum -y install go; sleep 1
cd go ; go get github.com/arangodb/go-driver ; cd -; sleep 1
cd go ; go mod init test ; go build ; ./test ; cd -; sleep 1
