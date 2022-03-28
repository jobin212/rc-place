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
- There's a 1 second timeout for placing tiles. Speed up testing by updating `updateLimit` in [client.go](../client.go) and [running locally](../README.md).
