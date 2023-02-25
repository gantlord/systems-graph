sudo docker run -p 8529:8529 -e ARANGO_NO_AUTH=1 arangodb/arangodb:3.10.3 > /tmp/oot 2>&1&
sleep 10
CONTAINER=`sudo docker ps | awk '{print $1}' | tail -n 1`
sudo docker cp data $CONTAINER:/tmp
sudo docker exec -it $CONTAINER arangorestore --server.authentication false /tmp/data/db
sudo docker ps
