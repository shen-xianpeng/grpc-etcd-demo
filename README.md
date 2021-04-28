

安装etcd
brew install etcd

安装etcd web浏览器工具 https://github.com/evildecay/etcdkeeper

启动etcd，使用3个不同端口运行三个服务端：

    go run server/main.go -Port=3000
    go run server/main.go -Port=3001
    go run server/main.go -Port=3002

启动客户端：
    
    go run client/main.go

可以看到，客户端使用轮询的方式对三个服务端进行请求，从而实现负载均衡。
