sudo yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo
sudo yum -y install gh
gh auth login
git config --global user.email robin.bruce@gmail.com
git config --global user.name "Robin Bruce"
