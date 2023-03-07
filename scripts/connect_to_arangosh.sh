CONTAINER=`sudo docker ps | awk '{print $1}' | tail -n 1`
sudo docker exec -it $CONTAINER arangosh --server.authentication
