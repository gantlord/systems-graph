CONTAINER=`sudo docker ps | awk '{print $1}' | tail -n 1`
sudo docker exec -it $CONTAINER rm -fr /tmp/db
sudo docker exec -it $CONTAINER arangodump --server.authentication false /tmp/db
sudo docker cp $CONTAINER:/tmp/db /tmp/
sudo chown -R `whoami` /tmp/db
cp -r /tmp/db data
