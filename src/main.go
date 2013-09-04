package main

import (
	"net/http"
	"html/template"
	"log"
	"os"
	"time"
	"fmt"
	"strings"
	"strconv"
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

type listenTaskInfo struct {
	servers string
	vport int
	dateTime string
}

// 主页请求处理函数
func getHomePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("../template/index.html")
	if err != nil {
		fmt.Println("Template Not Found!")
	}else {
		t.Execute(w, "")
	}
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
	messageParts := make([]string, 0, 2)
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
	return
}

// 获取haproxy listen任务列表
func getListenList(w http.ResponseWriter, r *http.Request){
	rows, err := db.Query("SELECT servers, vport, datetime FROM haproxymapinfo ORDER BY datetime DESC")
	if err != nil {
		fmt.Println(err)
	}

	var listenTasks = make([]listenTaskInfo, 0, 100)
	var servers string
	var vport int
	var dateTime string
	for rows.Next() {
		err = rows.Scan(&servers, &vport, &dateTime)
		append(listenTasks, listenTaskInfo{servers: servers, vport: vport, dateTime: dateTime})
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
	}
	t, err := template.ParseFiles("../templates/listenlist.html")
	t.Execute(w, listenTasks)
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
