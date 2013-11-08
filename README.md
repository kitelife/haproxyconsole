HAProxyConsole是一个简单的HAProxy负载均衡任务管理系统。由于HAProxy的负载均衡任务可能会很多，手动编辑配置文件非常不方便、不安全，所以实现一个友好的管理系统是非常必要的。

### 功能点：

1. TCP协议负载均衡任务的增删改、任务的列表展示；
2. 一键应用最新配置到主服务器或从服务器并重新HAProxy进程；
3. 修改一个配置项即可在JSON文件存储和数据库存储之间切换；
4. 内置小工具用于不同存储方式之间的数据转换；
5. 内嵌主从HAProxy自带数据统计页面，方便查看信息；
6. 分/不分业务端口自动分配、指定分配端口模式；
7. 内置配置文件正确性检查功能；等...

*基于Go语言标准库http实现自带Web server，一般情况不需再使用nginx/apache。*

### 使用场景（系统结构图）

![High Availability Load Balancer](https://raw.github.com/youngsterxyf/youngsterxyf.github.com/master/assets/uploads/pics/high-availability-load-balancer.png)

### 基本功能页面截图

![screenshot-1](https://raw.github.com/youngsterxyf/haproxyconsole/master/screenshot1.png)

![screenshot-2](https://raw.github.com/youngsterxyf/haproxyconsole/master/screenshot2.png)

### 配置：

conf目录下有4个文件：

- app.sql：如果选择以MySQL来存储，则执行该文件中的sql语句创建数据表
- app_conf.ini：该文件为haproxyconsole的主配置文件，使用之前阅读每个配置项的说明信息，按照说明修改配置。
- DB.json：该文件是在选择以JSON文件存储时，自动生成的存储文件。
- haproxy_conf_common.json：该json文件包含4项数据-“Global”、“Defaults”、“ListenStats”、“ListenCommon”，“Global”和“Default”对应HAProxy配置的Global和Defaults部分，“ListenStats”是启用HAProxy自带的数据统计页面，用户可能需要修改该功能启用的端口，“ListenCommon”是所有HAProxyConsole管理的TCP负载均衡任务在生成HAProxy配置listen块时的通用部分。

### 编译：

        cd src && go build main.go

### 使用：

启动HAProxyConsole（假设HAProxy部署在`/usr/local`目录下）：`cd /usr/local/haproxyconsole/bin && ./haproxyconsole &`，默认端口为9090，可使用选项`-p`来自定义端口，如：`cd /usr/local/haproxyconsole/bin && ./haproxyconsole -p 8080`（注意该端口不应在HAProxyConsole为HAProxy负载均衡任务自动分配的端口范围之内(10000-20000)）。

若需转换数据存储方式，则可通过内置工具来完成：`cd /usr/local/haproxyconsole/bin && ./haproxyconsole -t`，该工具完成的操作是：若StoreScheme设定为0，则从DBDriverName和DBDataSourceName 指定数据库的haproxymapinfo数据表中读取数据，转换成json格式存入FileToReplaceDB指定路径的JSON文件中。

### 缺点：
1. 没有手动编辑HAProxy配置文件灵活，没法对某些负载均衡任务做详细的定制。
2. 目前只实现了3层（TCP协议）负载均衡任务管理。

