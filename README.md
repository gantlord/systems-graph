create an AWS micro instance with an Amazon Linux image

run:

sudo yum -y install git

git clone https://github.com/gantlord/systems-graph.git

./setup.sh

GENERAL TIPS

there are helper scripts in the scripts directory - if you're in a "hey how do I?" type situation, cat'ing those files is probably a help as they are my go-to when I need to debug. 

scripts as subdivided in categories, so if you're running the setup.sh script a lot then you probably just need to run a sub-setup script, e.g. if your arangodb instance is up and running and you have a fresh checkout, you only need to run setup_go.sh probably
