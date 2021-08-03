# go-wrk

go-wrk is a simple webserver requester with the ability to test multiple sites at the same time.

It is build with [termui]https://(github.com/gizak/termui/v3) to create the data table.

For now simply git clone the project, build and run it with

```
go run main.go http://example.com https://google.com
```

This will start calling both sites and give you stats about their response time.

## Possible flags
-r to set the rate
This will try to call the sites 50 times a sec.
```
go run main.go -r 50 http://example.com
```

-c to set the number of callers.
You will need to increase this as you increase the rate as a caller will not call the site again before it have recieved a response.
```
go run main.go -r 50 -c 10 http://example.com
```

-max this wil simply set -r and -c to a large number and try to find the max requests the server can handle.
```
go run main.go -max http://example.com
```

## Installation
There is no prebuild binary for now, but you can build it yourself and place it in your `/usr/bin` folder to be able to use it without using being in the project folder.

```
go build -o go-wrk main.go
sudo mv go-wrk /usr/bin/
```
