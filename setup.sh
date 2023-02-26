ARANGO_VERSION=latest
sudo yum -y install docker
sudo systemctl start docker
sudo docker pull arangodb:$ARANGO_VERSION
sudo docker run -p 8529:8529 -e ARANGO_NO_AUTH=1 arangodb/arangodb:$ARANGO_VERSION > /tmp/oot 2>&1&
sleep 22
CONTAINER=`sudo docker ps | awk '{print $1}' | tail -n 1`
sudo docker cp data $CONTAINER:/tmp
sudo docker exec -it $CONTAINER arangorestore --server.authentication false /tmp/data/db
sudo docker ps
scripts/setup_go.sh
