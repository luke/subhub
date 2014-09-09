SubHub is pusher compatible pub/sub server written in Go that uses Redis for channels and state. 

It supports both web sockets and sockjs transports for fallback. 

In addition to pub/sub it also supports fetching tracking changes of redis keys.

Running the server.. 

cd ./cmd/subhub 
go build 
./subhub --help 

