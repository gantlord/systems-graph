sudo systemctl start docker
sudo docker ps | tail -n +2 | awk '{print $1}' | while read cont; do sudo docker stop $cont; done
sudo docker ps -a | tail -n +2 | awk '{print $1}' | while read cont; do sudo docker rm $cont; done
sudo docker images | tail -n +2 | awk '{print $3}' | while read img; do sudo docker rmi $img; done
