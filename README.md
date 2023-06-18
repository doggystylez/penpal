## 🖋️🤝

lightweight tendermint signing monitor

usage:
```
penpal -c <path to config.json>
```

config (generated on first run):
```
{
  "networks": [
    {
      "name": "Network1",
      "chain_id": "network-1",
      "address": "AAAABBBBCCCCDDDD",  # hex consensus address
      "rpcs": [
        "rpc1",
        "rpc2"
      ],
      "back_check": 5,                # number of blocks to check
      "interval": 15                  # check interval in minutes
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
      "key": "api_key"
      "chat": "chat_id"
    },
    "discord": {
        "webhook": "webhook_url"
    }
  }
}
```