# Facade

Hosting internal minecraft servers just for a bunch of friends who are not frequently online wastes a lot of resources: the vanilla / spigot / paper server implementations all suffer from moderate CPU usage when idling (20%-30% on my low-end homeserver), let alone consuming a decent amount of RAM. 

Facade aims to automatically shutdown (hibernate) the server when idle for a (configurable) fixed period of time, when a new player connects, it restarts the server and proxies all data to the server instance. 
This process is transparent to the players: the server status is always open to the players, and friendly messages are shown throughout the process (i.e., MOTD, connection timeout while starting the server etc.) to indicate the server is in hibernation.

## Features
* **Lightweight**: Written in Go, we strive to leave as little CPU / memory footprint as possible.

* **Automatic Shutdown and Restart**: Schedule to shutdown the server when idling for a period of time and automatically restart the server when needed.

* **Respond to Server List Ping**: When our minecraft server is hibernated due to inactivity, we still answer the server list ping packets that the client sends us, so that the clients still see our server as alive. This is done by implementing the server list ping protocol in our proxy server, the actual minecraft server is only restarted if a client tries to log in. 

* **Configurable**: Behaviors of facade are made configurable to fit the different needs from server owners.
