Simple Tcp Socket Server Demo
=========
This demo has three part.

### Server
* TCP server takes in any request text per line and send a query to an external API, until client send 'quit' or timed out.
* TCP server can accept multiple connections at the same time.
* TCP server default multiple connections limit is 100.
* As for the external API, the choice is yours, or even a mock.
* API request rate limit for the external API: `2` requests per second.
* TCP server has an http endpoint to show stats.

### Client
* A simple TCP client able to input command.

### API Server Mock
* A simple API Server Mock return you command.

## Usage

### TCP Server

```
go run demo.go
# or 
go run demo.go --server="localhost:30000"
``` 

### Client

```
go run demo.go --demoType=client
# or 
go run demo.go --demoType=client --server="localhost:30000"
# or use system command
nc localhost 30000
``` 

### API Server Mock

``` 
go run demo.go --demoType=apiserver
``` 

## How to play
* Start an TCP server
* Open client Connect to the server and type anything you will get `Server unreachable.`.
* Start the API server.
* Try to input any thing on client you will get responses.
* If client idle over 20 sec will lose connect. 
* If client send command to fast (over 2 requests per second.) will get `Too Many Requests.`.