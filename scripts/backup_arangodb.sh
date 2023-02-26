CONTAINER=`sudo docker ps | awk '{print $1}' | tail -n 1`
sudo docker exec -it $CONTAINER arangodump --server.authentication false --overwrite true /tmp/db
sudo docker cp $CONTAINER:/tmp/db /tmp/
sudo chown -R `whoami` /tmp/db
cp -r /tmp/db data
