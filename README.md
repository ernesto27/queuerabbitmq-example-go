# RABBITMQ QUEUE EXAMPLE 

This is an example project that uses the queue message broker rabbitMQ on a simil real scenario.

Architecture diagram.

![Alt text](architecture.jpg?raw=true "Title")

There are 3 services on the project

##### BACKEND
This services simulates some kind on api request from a new post client user.
First the service saves on a sqlite db a new item post.
After tries to connect to a rabbitmq service and sent a message with data from the new post.

##### EMAIL 
This service connect to a rabbitMQ service, gets all messages from the backend service and send a email to some user.

##### IMAGE 
This service connect to a rabbitMQ service gets all messages from the backend service and generate a new thumbnail image from a new item post.

##### QUEUESEND
This service is in charge of send again a message to a queue in case that some problem happened when the backend tries to connect with the queue.

##### DEV INSTALATION
Install the rabbitque service in a docker container

```sh
$ docker run -it -d  --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

Run the init db sqlite script only once.
```sh
$ go run initDB.go
```

Edit enviroment values on .env
```sh
$ cp .env-example .env
```

```
FROM_EMAIL= //your smt email
FROM_PASSWORD= // your smt password
TO_EMAIL= // email receiver
RABBIT_CONN= // rabbitmq connection
```

Simulate a new user post insert and sent a message to the queue
```sh
$ cd backend && go run backend.go
```

Run next commands on a new terminal to start listen for new messages.

Consumer email
```sh
$ cd email && emailConsumer.go
```

Consumer image
```sh
$ cd image && image.go
```

Service that get all backend posts that could not connect to queue and tries to sent message again.

```sh
$ cd queuesend && queuesend.go
```


#### Run process on background
If you want to run a go binary like a background process you can use systemd linux configuration.

Create file on path
```
/etc/systemd/system/emailconsumer.service
```

Add this content to that file.

```
[Unit]
Description=go rabbitmq service email
After=network.target
StartLimitIntervalSec=0
[Service]
Type=simple
Restart=always
RestartSec=1
User=ernesto
ExecStart=/home/username/code/go/rabbitmq/email/emailConsumer

[Install]
WantedBy=multi-user.target

```

Start service
```
$ systemctl start emailconsumer
```













