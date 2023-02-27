ARANGO_VERSION=latest
sudo yum -y install docker
sudo systemctl start docker
sudo docker pull arangodb:$ARANGO_VERSION
sudo docker run -p 8529:8529 -e ARANGO_NO_AUTH=1 arangodb/arangodb:$ARANGO_VERSION > /tmp/oot 2>&1&
sleep 22
