package main

import (
	"net/http"
	"html/template"
	"log"
	"os"
	//"os/exec"
	"time"
	"fmt"
	"strings"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// 声明全局变量
var logger *log.Logger
var db *sql.DB
var vip = "127.0.0.1"

// 定义applyVPort结果结构体
type applyResult struct {
	Success string
	Msg     string
}

// listen任务列表数据
type listenTaskInfo struct {
	Seq	  int
	Servers  template.HTML
	Vip      string
	Vport    int
	DateTime string
}

// 页面导航栏高亮数据
/*
type navHighlight struct {
	AddTask    string
	ListenList string
}
*/

//index页面，即添加任务页面，模板数据
/*
type indexData struct {
	Nav navHighlight
}
*/

// listenlist页面模板数据
type listenListData struct {
	ListenTaskList []listenTaskInfo
}

// 存储配置文件解析结果
type config struct {
	Global       []string
	Defaults     []string
	ListenStats  []string
	ListenCommon []string
}

// 主页请求处理函数
func getHomePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("../template/header.tmpl", "../template/index.tmpl", "../template/footer.tmpl")
	if err != nil {
		fmt.Println(err)
	}else {
		t.ExecuteTemplate(w, "index", nil)
	}
}

// 根据数据库数据重新生成HAProxy配置文件，并重启HAProxy
func rebuildHAProxyConf() {

	newConfigParts := make([]string,0, 50)

	bytes, err := ioutil.ReadFile("../conf/haproxy_conf.json")
	if err != nil {
		logger.Fatalln(err)
		return
	}
	var conf config
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		logger.Fatalln(err)
		return
	}
	newConfigParts = append(newConfigParts, strings.Join(conf.Global, "\n\t"))
	newConfigParts = append(newConfigParts, strings.Join(conf.Defaults, "\n\t"))

	rows, err := db.Query("SELECT servers, vport FROM haproxymapinfo ORDER BY vport ASC")
	if err != nil {
		logger.Println(err)
	}

	var servers string
	var vport int
	for rows.Next() {
		err = rows.Scan(&servers, &vport)
		serverList := strings.Split(servers, "-")
		backendServerInfoList := make([]string,0, 10)

		for i := 0; i < len(serverList); i++ {
			serverInfoPath := strings.Split(serverList[i], ":")
			backendServerInfoList = append(backendServerInfoList, fmt.Sprintf("server %s %s weight 3 check inter 2000 rise 2 fall 3", serverInfoPath[1], serverList[i]))
		}
		newConfigParts = append(newConfigParts, fmt.Sprintf("Listen-%d\n\tbind *:%d\n\t%s\n\n\t%s", vport, vport, strings.Join(conf.ListenCommon, "\n\t"), strings.Join(backendServerInfoList, "\n\t")))
	}
	newConfigParts = append(newConfigParts, strings.Join(conf.ListenStats, "\n\t"))
	newConfig := strings.Join(newConfigParts, "\n\n")
	haproxyConfFile, _ := os.OpenFile("../conf/haproxy.conf", os.O_CREATE | os.O_RDWR, 0666)
	defer haproxyConfFile.Close()
	haproxyConfFile.WriteString(newConfig)
	// 重启haproxy
	/*
	cmd := exec.Command("/usr/local/haproxy/sbin/haproxy -f /usr/local/haproxy/conf/haproxy.conf -st `/usr/local/haproxy/haproxy.pid`")
	err = cmd.Run()
	if err != nil {
		logger.Println(err)
	}
	*/
}

// 申请虚拟ip端口请求处理函数
func applyVPort(w http.ResponseWriter, r *http.Request) {
	servers := r.FormValue("servers")

	rows, err := db.Query("SELECT vport FROM haproxymapinfo ORDER BY vport DESC LIMIT 1")
	if err != nil {
		logger.Println(err)
	}
	var maxiumVPort int
	for rows.Next() {
		err = rows.Scan(&maxiumVPort)
		if err == nil {
			break
		}
	}
	if maxiumVPort == 0 {
		maxiumVPort = 10000
	}
	vportToAssign := maxiumVPort + 1
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec("INSERT INTO haproxymapinfo (servers, vport, datetime) VALUES (?, ?, ?)", servers, vportToAssign, now)
	if err != nil {
		logger.Println(err)
	}
	messageParts := make([]string,0, 2)
	messageParts = append(messageParts, vip)
	messageParts = append(messageParts, strconv.Itoa(vportToAssign))
	message := strings.Join(messageParts, ":")

	result := applyResult{
		Success: "true",
		Msg: message,
	}
	rt, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(rt))
	go rebuildHAProxyConf()
	return
}

// 获取haproxy listen任务列表
func getListenList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT servers, vport, datetime FROM haproxymapinfo ORDER BY datetime DESC")
	if err != nil {
		fmt.Println(err)
	}

	var listenTasks = make([]listenTaskInfo,0, 100)
	var servers string
	var vport int
	var dateTime string
	seq := 1
	for rows.Next() {
		err = rows.Scan(&servers, &vport, &dateTime)
		lti := listenTaskInfo{
			Seq: seq,
			Servers: template.HTML(strings.Join(strings.Split(servers, "-"), "<br />")),
			Vip: vip,
			Vport: vport,
			DateTime: dateTime,
		}
		listenTasks = append(listenTasks, lti)
		seq = seq + 1
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
	}
	Lld := listenListData{
		ListenTaskList: listenTasks,
	}

	t, err := template.ParseFiles("../template/header.tmpl", "../template/listenlist.tmpl", "../template/footer.tmpl")
	if err != nil {
		fmt.Println(err)
	}
	t.ExecuteTemplate(w, "listenlist", Lld)
	return
}

// 日志初始化函数
func getLogger() (logger *log.Logger) {
	os.Mkdir("../log/", 0666)
	logFile, err := os.OpenFile("../log/HAProxyConsole.log", os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger = log.New(logFile, "\r\n", log.Ldate | log.Ltime | log.Lshortfile)
	return
}

func main() {

	var err error
	logger = getLogger()

	// 数据库连接初始化
	db, err = sql.Open("mysql", "root:06122553@tcp(localhost:3306)/haproxyconsole?charset=utf8")
	if err != nil {
		logger.Fatalln(err)
		os.Exit(1)
	}
	defer db.Close()

	// 请求路由
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static/"))))
	http.HandleFunc("/applyvport", applyVPort)
	http.HandleFunc("/listenlist", getListenList)
	http.HandleFunc("/", getHomePage)

	// 启动http服务
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		logger.Fatalln("ListenAndServe: " , err)
	}
}
