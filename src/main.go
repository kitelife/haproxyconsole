package main

import (
	"applicationDB"
	"config"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sshoperation"
	"strconv"
	"strings"
	"time"
	"tools"
)

// 声明全局变量
var logger *log.Logger
var db applicationDB.DB
var appConf config.ConfigInfo

// 状态结果结构体
type statusResult struct {
	Success string
	Msg     string
}

// 主页请求处理函数
func getHomePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("../template/header.tmpl", "../template/index.tmpl", "../template/footer.tmpl")
	if err != nil {
		fmt.Println(err)
	} else {
		t.ExecuteTemplate(w, "index", nil)
	}
}

// 根据数据库数据重新生成HAProxy配置文件
func rebuildHAProxyConf() {

	// 存储配置文件解析结果
	type config struct {
		Global       []string
		Defaults     []string
		ListenStats  []string
		ListenCommon []string
	}

	newConfigParts := make([]string, 0, 50)

	bytes, err := ioutil.ReadFile("../conf/haproxy_conf_common.json")
	//fmt.Println(string(bytes))
	if err != nil {
		return
	}
	var conf config
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return
	}
	newConfigParts = append(newConfigParts, strings.Join(conf.Global, "\n\t"))
	newConfigParts = append(newConfigParts, strings.Join(conf.Defaults, "\n\t"))
	newConfigParts = append(newConfigParts, strings.Join(conf.ListenStats, "\n\t"))

	dataList, err := db.QueryNewConfData()
	if err != nil {
		logger.Println(err)
	}

	taskNum := len(dataList)
	for index := 1; index < taskNum; index++ {
		task := dataList[index]
		serverList := strings.Split(task.Servers, "-")
		backendServerInfoList := make([]string, 0, 10)
		serverNum := len(serverList)
		for i := 0; i < serverNum; i++ {
			backendServerInfoList = append(backendServerInfoList, fmt.Sprintf("server %s %s weight 3 check inter 2000 rise 2 fall 3", serverList[i], serverList[i]))
		}
		listenCommon := conf.ListenCommon
		if task.LogOrNot == 1 {
			logDirective := "option tcplog\n\tlog global\n\t"
			listenCommon = append(listenCommon, logDirective)
		}
		newConfigParts = append(newConfigParts, fmt.Sprintf("listen Listen-%d\n\tbind *:%d\n\t%s\n\n\t%s", task.VPort, task.VPort, strings.Join(listenCommon, "\n\t"), strings.Join(backendServerInfoList, "\n\t")))
	}

	newConfig := strings.Join(newConfigParts, "\n\n")
	// 必须使用os.O_TRUNC来清空文件
	haproxyConfFile, err := os.OpenFile(appConf.NewHAProxyConfPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer haproxyConfFile.Close()
	haproxyConfFile.WriteString(newConfig)
	return
}

// 申请虚拟ip端口请求处理函数
func applyVPort(w http.ResponseWriter, r *http.Request) {

	servers := r.FormValue("servers")
	comment := r.FormValue("comment")
	logOrNot := r.FormValue("logornot")

	rows, err := db.QueryVPort()
	if err != nil {
		logger.Println(err)
	}
	rowNum := len(rows)
	/*
		虚拟ip端口分配算法
	*/
	// 端口占用标志位数组
	var vportToAssign int

	if rowNum == 0 {
		vportToAssign = 10000
	} else {
		var portSlots [10000]bool
		for index := 0; index < 10000; index++ {
			portSlots[index] = false
		}
		for index := 0; index < rowNum; index++ {
			port := rows[index]
			portSlots[port-10000] = true
		}
		maxiumVPort := rows[rowNum-1]
		vportToAssign = maxiumVPort + 1
		if (rowNum + 9999) < maxiumVPort {
			boundary := maxiumVPort - 9999
			for index := 0; index < boundary; index++ {
				if portSlots[index] == false {
					vportToAssign = index + 10000
					break
				}
			}
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	logornot, _ := strconv.Atoi(logOrNot)
	//fmt.Printf("servers: %s, vportToAssign: %d, comment: %s, logornot: %d, now: %s", servers, vportToAssign, comment, logornot, now)
	err = db.InsertNewTask(servers, vportToAssign, comment, logornot, now)
	if err != nil {
		logger.Println(err)
	}
	messageParts := make([]string, 0, 2)
	messageParts = append(messageParts, appConf.Vip)
	messageParts = append(messageParts, strconv.Itoa(vportToAssign))
	message := strings.Join(messageParts, ":")

	result := statusResult{
		Success: "true",
		Msg:     message,
	}
	rt, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(rt))
	return
}

// 获取haproxy listen任务列表
func getListenList(w http.ResponseWriter, r *http.Request) {

	// listen任务列表数据
	type listenTaskInfo struct {
		Seq      int
		Id       int
		Servers  template.HTML
		Vip      string
		Vport    int
		Comment  string
		LogOrNot int
		DateTime string
	}
	// listenlist页面模板数据
	type listenListData struct {
		ListenTaskList []listenTaskInfo
	}

	rows, err := db.QueryForTaskList()
	if err != nil {
		logger.Println(err)
	}

	var listenTasks = make([]listenTaskInfo, 0, 100)
	seq := 1
	rowNum := len(rows)
	for index := 0; index < rowNum; index++ {
		row := rows[index]
		lti := listenTaskInfo{
			Seq:      seq,
			Id:       row.Id,
			Servers:  template.HTML(strings.Join(strings.Split(row.Servers, "-"), "<br />")),
			Vip:      appConf.Vip,
			Vport:    row.VPort,
			Comment:  row.Comment,
			LogOrNot: row.LogOrNot,
			DateTime: row.DateTime,
		}
		listenTasks = append(listenTasks, lti)
		seq = seq + 1
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

// 删除任务
func delListenTask(w http.ResponseWriter, r *http.Request) {

	success := "true"
	msg := "已成功删除"

	taskId := r.FormValue("taskid")
	id, _ := strconv.Atoi(taskId)
	result, err := db.DeleteTask(id)
	if err != nil {
		logger.Fatalln(err)
		success = "false"
		msg = "删除数据出错！"
	}
	rowsAffected, err := result.RowsAffected()
	if rowsAffected != 1 {
		success = "false"
		msg = fmt.Sprintf("数据删除有问题，删除了%d条", rowsAffected)
	}
	rt, _ := json.Marshal(statusResult{Success: success, Msg: msg})
	fmt.Fprintf(w, string(rt))
	return
}

// 编辑任务
func editTask(w http.ResponseWriter, r *http.Request) {

	success := "true"
	msg := "更新成功！"

	servers := r.FormValue("servers")
	comment := r.FormValue("comment")
	logornot := r.FormValue("logornot")
	id := r.FormValue("id")
	now := time.Now().Format("2006-01-02 15:04:05")
	logOrNot, _ := strconv.Atoi(logornot)
	taskId, _ := strconv.Atoi(id)
	err := db.UpdateTaskInfo(servers, comment, logOrNot, now, taskId)
	if err != nil {
		logger.Println(err)
		success = "false"
		msg = "数据存储操作出现问题!"
	}
	result := statusResult{
		Success: success,
		Msg:     msg,
	}
	rt, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(rt))
	return
}

// 应用新HAProxy配置文件
func applyConf(w http.ResponseWriter, r *http.Request) {

	success := "true"
	msg := "成功应用！"

	target := r.FormValue("target")
	rebuildHAProxyConf()
	if target == "master" {
		// 重启haproxy
		cmd := fmt.Sprintf("cp %s %s && %s", appConf.NewHAProxyConfPath, appConf.MasterConf, appConf.MasterRestartScript)
		cmdToRun := exec.Command(cmd)
		err := cmdToRun.Run()
		if err != nil {
			logger.Println(err)
			success = "false"
			msg = fmt.Sprintf("应用失败！%s", err.Error())
		}
	} else {
		err := sshoperation.ScpHaproxyConf(appConf)
		if err != nil {
			success = "false"
			msg = fmt.Sprintf("应用失败！%s", err.Error())
		}
	}
	result := statusResult{
		Success: success,
		Msg:     msg,
	}
	rt, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(rt))
	return
}

// 展示HAProxy自带的数据统计页面
func statsPage(w http.ResponseWriter, r *http.Request) {

	type StatsPageData struct {
		StatsUrl string
	}

	target := r.FormValue("target")
	url := appConf.MasterStatsPage
	if target == "slave" {
		url = appConf.SlaveStatsPage
	}
	t, _ := template.ParseFiles("../template/header.tmpl", "../template/statspage.tmpl")
	spd := StatsPageData{
		StatsUrl: url,
	}
	t.ExecuteTemplate(w, "statspage", spd)
	return
}

// 日志初始化函数
func getLogger() (logger *log.Logger) {
	os.Mkdir("../log/", 0666)
	logFile, err := os.OpenFile("../log/HAProxyConsole.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger = log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

func main() {

	var err error
	logger = getLogger()
	appConf, _ = config.ParseConfig("../conf/app_conf.ini")

	port := flag.String("p", "9090", "port to run the web server")
	init := flag.Bool("i", false, "init to create the haproxymapinfo table in database")
	toolMode := flag.Bool("t", false, "run this program as a tool to export data from database to json or from json to database")

	flag.Parse()
	if *init {
		// 初始化创建数据表haproxymapinfo
		err := tools.InitDataTable(appConf)
		tools.CheckError(err)
	} else {
		if *toolMode {
			// 数据转换存储方式
			err := tools.StorageTransform(appConf)
			tools.CheckError()
		} else {
			// 存储连接初始化
			db, err = applicationDB.InitStoreConnection(appConf)
			if err != nil {
				logger.Fatalln(err)
				os.Exit(1)
			}
			defer db.Close()

			// 请求路由
			http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static/"))))
			http.HandleFunc("/applyvport", applyVPort)
			http.HandleFunc("/edittask", editTask)
			http.HandleFunc("/listenlist", getListenList)
			http.HandleFunc("/dellistentask", delListenTask)
			http.HandleFunc("/applyconf", applyConf)
			http.HandleFunc("/statspage", statsPage)
			http.HandleFunc("/", getHomePage)

			// 启动http服务
			err = http.ListenAndServe(":"+*port, nil)
			if err != nil {
				logger.Fatalln("ListenAndServe: ", err)
			}
		}
	}
}
