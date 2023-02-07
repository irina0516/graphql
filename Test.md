\\\\.\\pipe\\galaxy.ipc
/home/blockchain/.galaxy/galaxy.ipc
/root/axis-graphql/build/apiserver -cfg ./config.json

sudo /home/ubuntu/graphql/apiserver -cfg /home/ubuntu/graphql/config.json
cd /home/ubuntu/binary

sudo /home/ubuntu/axis-binary/graphql-ubuntu-20.04-lts -cfg /home/ubuntu/axis-binary/config.json


cd /etc/systemd/system
sudo vi graphql.service

[Unit]
Description=graphql service
After=network.target
[Service]
Type=simple
User=ubuntu
ExecStart=sudo /home/ubuntu/graphql/apiserver -cfg /home/ubuntu/graphql/config.json
Restart=on-failure
[Install]
WantedBy=multi-user.target

sudo systemctl enable graphql
sudo systemctl restart graphql
sudo systemctl status graphql