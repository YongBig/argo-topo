# argo-topo

Detach argocd topology function demo

Redis 同步保存topo信息

## 接口

查询接口 curl --location --request GET '127.0.0.1:8081/topologys/{namespace}/{name}

参数

namesapce:  部署 application 所有namespace 名称 name:       部署 application 名称

## 启动：

1、启动redis

```shell
cd ./deployment 
# 启动redis 客户端 与 服务端
# redis 127.0.0.1:6379  
# redis ui 127.0.0.1:7843
docker-compose up -d 
```

2、启动topo

```shell
go run main.go
```
