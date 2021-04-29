package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-kit-etcd-demo/lib/logger"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	installationProto "go-kit-etcd-demo/lib/proto/installation"
	registryProto "go-kit-etcd-demo/lib/proto/registry"
	installationServer "go-kit-etcd-demo/server/installation"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
)

const (
	namespace = "72changes"
	VERSION   = "0.0.1"
)

var (
	panicRecoveryFunc grpc_recovery.RecoveryHandlerFunc
)

var host = "127.0.0.1" //服务器主机
var (
	Port        = flag.Int("Port", 3000, "listening port")                           //服务器监听端口
	ServiceName = flag.String("ServiceName", "grpc_service", "service name")         //服务名称
	EtcdAddr    = flag.String("EtcdAddr", "127.0.0.1:2379", "register etcd address") //etcd的地址
)
var cli *clientv3.Client

//将服务地址注册到etcd中
func register(etcdAddr, serviceName, serverAddr string, version string, ttl int64) error {
	var err error

	if cli == nil {
		//构建etcd client
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(etcdAddr, ";"),
			DialTimeout: 15 * time.Second,
		})
		if err != nil {
			logger.ErrorMsg("连接etcd失败：%s", err.Error())
			return err
		}
	}

	//与etcd建立长连接，并保证连接不断(心跳检测)
	ticker := time.NewTicker(time.Second * time.Duration(ttl))
	go func() {
		key := "/" + namespace + "/" + serviceName + "/" + serverAddr
		for {
			resp, err := cli.Get(context.Background(), key)
			//fmt.Printf("resp:%+v\n", resp)
			if err != nil {
				logger.ErrorMsg("获取服务地址失败：%s", err.Error())
			} else if resp.Count == 0 { //尚未注册
				err = keepAlive(serviceName, serverAddr, VERSION, ttl)
				if err != nil {
					logger.ErrorMsg("保持连接失败：%s", err.Error())
				}
			}
			<-ticker.C
		}
	}()

	return nil
}

//保持服务器与etcd的长连接
func keepAlive(serviceName, serverAddr string, version string, ttl int64) error {
	//创建租约
	leaseResp, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		logger.ErrorMsg("创建租期失败：%s", err.Error())
		return err
	}

	//将服务地址注册到etcd中
	key := "/" + namespace + "/" + serviceName + "/" + serverAddr
	val := registryProto.GrpcNodeInfo{
		Addr:    serverAddr,
		Version: version,
	}

	valBytes, _ := json.Marshal(val)
	_, err = cli.Put(context.Background(), key, string(valBytes), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		logger.ErrorMsg("注册服务失败：%s", err.Error())
		return err
	}

	//建立长连接
	ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		logger.ErrorMsg("建立长连接失败：%s", err.Error())
		return err
	}

	//清空keepAlive返回的channel
	go func() {
		for {
			<-ch
		}
	}()
	return nil
}

//取消注册
func unRegister(serviceName, serverAddr string) {
	if cli != nil {
		key := "/" + namespace + "/" + serviceName + "/" + serverAddr
		cli.Delete(context.Background(), key)
	}
}

func main() {
	flag.Parse()
	logger.InitLogger()
	//监听网络
	serverAddr := fmt.Sprintf("%s:%d", host, *Port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		logger.Error("监听网络失败：", err)
		return
	}
	defer listener.Close()

	panicRecoveryFunc = func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(panicRecoveryFunc),
	}
	//创建grpc句柄
	srv := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_recovery.StreamServerInterceptor(opts...),
		),
	)
	defer srv.GracefulStop()

	//将installationServer结构体注册到grpc服务中
	installSrv, errs := installationServer.NewInstallationServer(
		installationServer.Addr(serverAddr),
	)
	if errs != nil {
		logger.ErrorMsg("installationServer init error %#v", errs)
	}
	installationProto.RegisterInstallationServer(
		srv,
		installSrv,
	)
	reflection.Register(srv)

	//将服务地址注册到etcd中
	logger.InfoMsg("grpc server address: %s", serverAddr)
	register(*EtcdAddr, *ServiceName, serverAddr, VERSION, 5)

	//关闭信号处理
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		unRegister(*ServiceName, serverAddr)
		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	//监听服务
	err = srv.Serve(listener)
	if err != nil {
		logger.ErrorMsg("监听异常：%s", err.Error())
		return
	}
}
