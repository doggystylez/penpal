## 🖋️🤝

lightweight tendermint signing monitor

## usage
generate config
```
penpal -c </path/to/config.json> -init
```
run
```
penpal -c </path/to/config.json>
```

## health check
multiple instances can be run to monitor each other and alert if any instance is malfunctioning or unavailable. currently, this uses a http server which listens on the designated port with no authentication other than checking the header, so use firewall rules to only allow access to this port from the other instances

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
			"back_check": 5, # number of blocks to check
			"interval": 15   # check interval in minutes
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
			"interval": 15
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
		"interval": 1, # health check interval in hours
		"port": "8080", # health listen port
		"nodes": [
			"http://192.168.1.1:8080" # addresses of other instances to run health checks on
		]
	}
}
```
