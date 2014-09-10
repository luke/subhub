SubHub is pusher compatible pub/sub server written in Go that uses Redis for channels and state. 

It supports both web sockets and sockjs transports for fallback. 

In addition to pub/sub it also supports fetching tracking changes of redis keys.

Running the server.. 

cd ./cmd/subhub 
go build 
./subhub --help 

See cmd/subhub/web/pusher.html for javascript client. 

Note to run locally you will need to add.. 

127.0.0.1 ws.screencloud.io
127.0.0.1 sockjs.screencloud.io
127.0.0.1 js.screencloud.io

Object channels

var foo1 = pusher.subscribe("object-foo1"); 
foo1.bind("load",function(data){console.log("loaded", data);}); 
foo1.bind("change",function(data){console.log("changed!", data);});

