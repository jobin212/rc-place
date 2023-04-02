# RC-Place Bots

## Setup
Create a personal access token on [https://www.recurse.com/settings/apps](https://www.recurse.com/settings/apps).  Set PERSONAL_ACCESS_TOKEN in your environmental variables (see .env.example).

```shell
# Load your environmental variables after setting them.
🎨 source .env.example

# Run shell bot
🎨 ./rc.sh

# Run python bot
🎨 python3 rc.py
```

## Tips
- There's a 1 second timeout for placing tiles. Speed up testing by updating `updateLimit` in [client.go](../client.go) and [running locally](../README.md#build-and-run).
- Take a look at the [API](../README.md#rest-api)

### How to get an appendonly.aof file from fly Redis
1. Set up wireguard for fly access using [Kurt's advice here](https://community.fly.io/t/ssh-connection-to-an-instance/834)
2. Get your app ip using `flyctl ips private`
3. `scp -6 root@'[IPV6_ADDRESS]':/tmp/appendonly.aof ./appendonly.aof`
