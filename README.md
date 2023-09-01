## 🖋️🤝 Penpal Monitor

lightweight tendermint signing monitor
## build
```
git clone https://github.com/cordtus/penpal.git

cd penpal

go build ./cmd/penpal
```
## generate and set config
```
./penpal -init

nano config.json
```


## set up systemd service
save the following as `/etc/systemd/system/penpal.service`
```
[Unit]
Description=Penpal Monitor
After=network.target
[Service]
Type=simple
User=<user>
ExecStart=<path/to>/penpal -c <path/to>/config.json
Restart=always
RestartSec=2
[Install]
WantedBy=multi-user.target
```
```
systemctl start penpal

systemctl enable penpal

journalctl -u penpal.service -f -ocat
```

## heavily modified 
This is a very hacked up fork and is no longer a personal alerting tool. Its a very basic full-chain validator alerting tool for discord/telegram bot. At least 50% functional, half of the time. 