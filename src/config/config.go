package config

/*
主从HAProxy的VIP
*/
var Vip = "192.168.2.201"


/********************************************************************************
 ********************************************************************************/


/*
主HAProxy配置文件的路径
*/
var MasterConf = "/usr/local/haproxy/conf/haproxy.conf"
/*
主HAProxy重启脚本的路径
*/
var MasterRestartScript = "/usr/local/haproxy/restart_haproxy.sh"


/********************************************************************************
 ********************************************************************************/


/*
从HAProxy配置文件的路径
*/
var SlaveConf = "/usr/local/haproxy/conf/haproxy.conf"
/*
从HAProxy重启脚本的路径
*/
var SlaveRestartScript = "/usr/local/haproxy/restart_haproxy.sh"
/*
从HAProxy机器远程登录：服务器ip:port,用户名,密码
*/
var SlaveServer = "192.168.2.193:36000"
var SlaveRemoteUser = "root"
var SlaveRemotePasswd = "first2012++"


/********************************************************************************
 ********************************************************************************/


/*
根据负载均衡任务数据生成的HAProxy新配置文件存放路径
*/
var NewHAProxyConfPath = "../conf/haproxy_new.conf"


/********************************************************************************
 ********************************************************************************/


/*
数据存储方式：数据库(DB,以0表示)与文件(FILE,以1表示)两种
*/
var StoreScheme = 0

/*
MySQL数据库连接信息
*/
var DBDriverName = "mysql"
var DBDataSourceName = "root:06122553@tcp(127.0.0.1:3306)/haproxyconsole?charset=utf8"

/*
若采用文件来存储负载均衡任务数据，则需指定该文件路径
负载均衡任务数据文件路径
*/
var FileToReplaceDB = "../conf/DB.json"
