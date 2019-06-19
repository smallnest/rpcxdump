# rpcxdump

A tcpdump-like tool to capture rpcx tcp packets for debugging [rpcx services](https://github.com/smallnest/rpcx).

You **do not** need to modify/restart any existed services. One thing what you need is to copy this tool to the server and begin to dump.

It is convenient to debug communication between rpcx services and clients.

run it as :

```sh
./xdump -c 127.0.0.1:8972 -p -color
```

![](snapshoot.png)