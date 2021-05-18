# wheatDFS 

### 介绍

wheatDFS是一个基于GoRpc封装的，快速、简单、扩展性良好的分布式文件管理系统。

### 特性

- gorpc封装，友好的Api
- 稳定运行，高扩展性
- 提供go原生客户端（[WheatClient](https://gitee.com/timedb/wheatClient/)）
- 提供HttpAPi连接集群
- 支持断点续传
- 支持自动同步
- 大文件自动分割
- 文件令牌，统一管理大小文件
- Tracker（Leader）自动继承
- tracker集群负载均衡，无需Nginx等服务器



### 软件架构

![设计图2.0.2](https://gitee.com/timedb/picgo-imig/raw/master/images/%E8%AE%BE%E8%AE%A1%E5%9B%BE2.0.2.png)



### 安装教程

##### go语言安装

运行需要go语言环境, Go 语言环境安装[go中文网](https://studygolang.com) 搜索安装教程，需要安装**1.13**以上版本。

##### go proxy配置

打开你的终端并执行

```shell
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

完成。



##### wheatDFS安装 (linux)

```shell
cd ~
mkdir wheatDFS
git clone https://github.com/timedb/wheatDFS.git  # 下载仓库
cd goDFS
go build -o dfs
sudo cp dfs /bin  # 拷贝编译文件到bin下
```

##### wheatDFS安装 (windos)

```shell
cd /d D:/wheatDFS
mkdir wheatDFS
git clone https://github.com/timedb/wheatDFS.git  # 下载仓库
cd goDFS
go build -o dfs.exe  # 编译文件
```

配置环境变量：把dfs.exe所在地址配置到环境变量中



##### 查看帮助以及是否安装成功 

```shell
dfs -h  # 查看帮助

# 输出结果，表示安装成功
Usage of D:\gotest\dfs.exe:
  -conf string
        Specifies the configuration file to start the service (default "./wheatDFS.ini")
  -nc string
        Initializes a default configuration file.The output address needs to be specified
  -type string
        Use type to specify a service type. The value can only be tracker or storage
```



### 快速使用

##### 1. 创建配置文件

```shell
# 使用以下命令生成配置文件
# ./wheatDFS.ini 生成配置文件的地址

dfs -nc=./wheatDFS.ini
```

配置文件说明, 关键部分

```toml

version = "2.0.1"
debug = true  # debug为false时启动日志信息保存，否者打印到控制台

[tracker]
# this parameter is conforming leader's ip
ip = "192.168.31.173"   # 主机的IP，默认为创建该配置文件的服务器IP

# this parameter is conforming leader's port
port = "5590"  # 初始主机的端口

# the database of the fast-upload path 
esotericaPath = "./wheatDFS.eso"  # 保存上传的文件信息


[storage]
# this parameter is making the storage path in your server
groupPath = "D:/gotest/group"   # 重要，配置文件服务器保存文件的地址  


[log]
# the path of log database
logPath = "./log.db"  # debug为false时保存日志的地址

```

修改配置文件debug为false

##### 2. 启动主机Tracker

```shell
# 第一个启动的tracker必须是主机tracker（配置文件中 ip 对应的主机）
# 使用以下命令启动tracker
dfs -type=tracker -conf=./wheatDFS.ini
# conf可以不指定，默认为./wheatDFS.ini

# 出现 bind host 时代表启动成功
```

##### 3. 启动从机Tracker

wheatDFS支持Tracker集群启动，且信息自动同步，使用以下命令接入多个Tracker

```shell
# 保证配置文件[tracker]中的ip和port相同的配置文件启动会被认为是同一个wheatDFS服务

# 以下使用e1，e2表示不同的服务器

# ---e1---
# 拷贝主机Trakcer生成的配置文件到e1中, 地址为./wheatDFS.ini
# 使用以下命令启动tracker
dfs -type=tracker -conf=./wheatDFS.ini
# 出现 bind host 时代表启动成功


# ---e2---
# 拷贝主机Trakcer生成的配置文件到e2中, 地址为./wheatDFS.ini
# 使用以下命令启动tracker
dfs -type=tracker -conf=./wheatDFS.ini
# 出现 bind host 时代表启动成功
```

##### 4. Storage接入

```shell
# ---e1---
# 拷贝主机Trakcer生成的配置文件到e1中, 地址为./wheatDFS.ini
# 创建一个文件保存目录
mkdir /home/e1/group 
# 修改配置文件./wheatDFS.ini中的groupPath = "/home/e1/group" 
# 使用以下命令启动storage
dfs -type=storage -conf=./wheatDFS.ini
# 出现 bind host 时代表启动成功


# ---e2---
# 拷贝主机Trakcer生成的配置文件到e2中, 地址为./wheatDFS.ini
# 创建一个文件保存目录
mkdir /home/e2/group 
# 修改配置文件./wheatDFS.ini中的groupPath = "/home/e2/group" 
# 使用以下命令启动storage
dfs -type=storage -conf=./wheatDFS.ini
# 出现 bind host 时代表启动成功
```

使用以上步骤启动了一个3个Tracker, 2个Storage 的wheatDFS服务,我们可以使用[wheatClient](https://gitee.com/timedb/wheatClient/)来进行服务测试。