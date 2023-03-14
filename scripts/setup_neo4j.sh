NEO4J_VERSION=latest
sudo yum -y install docker
sudo systemctl start docker
sudo docker pull neo4j:$NEO4J_VERSION
sudo docker run \
    --publish=7474:7474 --publish=7687:7687 \
    --volume=$HOME/neo4j/data:/data \
    --env=NEO4J_AUTH=none\
    neo4j > /tmp/oot 2>&1&
sleep 22
