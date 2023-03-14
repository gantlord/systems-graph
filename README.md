BASIC SETUP

1. create an AWS micro instance with an Amazon Linux image, connect to it, then run:
2. sudo yum -y install git
3. git clone https://github.com/gantlord/systems-graph.git
4. cd systems-graph; ./setup.sh

NEO4J BROWSER

To connect to the Browser, go to the instance summary for your AWS instance:

1. Go the Security tab and edit the "Inbound rules" add a custom TCP rule for ports 7474 and 7687
2. Copy the "Public IPv4 DNS" entry from the "Instance summary" - to get something like "ec2-34-204-101-196.compute-1.amazonaws.com"
3. Add "http://$URL:7474" to what you got above and paste it into your browser window, e.g. http://ec2-34-204-101-196.compute-1.amazonaws.com:7474

GENERAL TIPS

there are helper scripts in the scripts directory - if you're in a "hey how do I?" type situation, cat'ing those files is probably a help as they are my go-to when I need to debug. 

scripts as subdivided in categories, so if you're running the setup.sh script a lot then you probably just need to run a sub-setup script, e.g. if your neo4j instance is up and running and you have a fresh checkout, you only need to run setup_go.sh probably

BACKGROUND READING

To get across the basic concepts and learn how to craft queries, you can follow the courses on: https://graphacademy.neo4j.com
