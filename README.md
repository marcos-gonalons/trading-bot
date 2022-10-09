Setup local environment
---
1. Install docker and docker-compose
2. Build docker image with `make build`
3. Run the bot with `make up`, or if you want to debug it, use `make up-debug`  
3.1. In order to debug it, it's necessary to attach to the docker container using port 2345  
In debug mode, the bot doesn't start until the debug process is started  
For vscode configuration, this is the launch.json file  
Then you can start the bot process with F5
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Connect to server",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "remotePath": "/TradingBot",
            "port": 2345,
            "host": "127.0.0.1"
        }
    ]
}
```


