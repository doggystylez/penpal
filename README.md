## 🖋️🤝 Penpal Monitor

lightweight tendermint signing monitor
## build
```
git clone https://github.com/doggystylez/penpal.git

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

## health check
multiple instances can be run to monitor each other and alert if any instance is unavailable. currently, this uses a http server which listens on the designated port with no authentication other than checking the header, so use firewall rules to only allow access to this port from the other instances

for example, running instance0 on 10.0.0.0:1000 and instance1 on 10.0.0.1:1001:

instance0 config
```
<...>
"nodes": ["http://10.0.0.1:1001"]
<...>
```
instance1 config
```
<...>
"nodes": ["http://10.0.0.0:1000"]
<...>
```

## config (JSON)
```
{
	"networks": [{
			"name": "Network1",
			"chain_id": "network-1",
			"address": "AAAABBBBCCCCDDDD",
			"rpcs": [
				"rpc1",
				"rpc2"
			],
			"back_check": 10,        # number of blocks to check
			"alert_threshold": 5,    # minimum of missed blocks to alert
			"interval": 15,          # check interval, in minutes
			"stall_time": 30         # alert if latest block timestamp is older than, in minutes
		},
		{
			"name": "Network2",
			"chain_id": "network-2",
			"address": "AAAABBBBCCCCDDDD",
			"rpcs": [
				"rpc1",
				"rpc2"
			],
			"back_check": 5,
            "alert_threshold": 1,
			"interval": 15,
			"stall_time": 30
		}

	],
	"notifiers": {
		"telegram": {
			"key": "api_key",
			"chat_id": "chat_id"
		},
		"discord": {
			"webhook": "webhook_url"
		}
	},
	"health": {
		"interval": 1,                # health check interval, in hours
		"port": "8080",               # health listen port
		"nodes": [
			"http://192.168.1.1:8080" # addresses of other instances to run health checks on
		]
	}
}
```
