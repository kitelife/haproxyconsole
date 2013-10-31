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
	type BusinessType struct {
		Index        int
		BusinessName string
	}
	type DataToRender struct {
		// Mode为0表示不分业务，为1表示分业务
		Mode             int
		BusinessTypeList []BusinessType
	}
	mode := 0
	businessTypeList := make([]BusinessType,0, 10)
	//fmt.Printf("%T\n", appConf.BusinessList)
	if appConf.BusinessList != "" {
		mode = 1
		businesses := strings.Split(appConf.BusinessList, ";")
		for index, item := range businesses {
			businessTypeList = append(businessTypeList, BusinessType{Index: index, BusinessName: strings.Split(item, ",")[0]})
		}
	}
	//fmt.Println(mode)
	t, err := template.ParseFiles("../template/header.tmpl", "../template/index.tmpl", "../template/footer.tmpl")
	if err != nil {
		fmt.Println(err)
	}

	dataToRender := DataToRender{Mode: mode, BusinessTypeList: businessTypeList}
	t.ExecuteTemplate(w, "index", dataToRender)
	return
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

	newConfigParts := make([]string,0, 50)

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
	for index := 0; index < taskNum; index++ {
		task := dataList[index]
		serverList := strings.Split(task.Servers, "-")
		backendServerInfoList := make([]string,0, 10)
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
	haproxyConfFile, err := os.OpenFile(appConf.NewHAProxyConfPath, os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer haproxyConfFile.Close()
	haproxyConfFile.WriteString(newConfig)
	return
}

// 自动分配端口算法
func autoAssignPort(firstPort int, lastPort int, assignedBiggest int, portSlots []bool) (vportToAssign int, noAvailablePort bool) {

	vportToAssign = -1
	noAvailablePort = false
	if assignedBiggest < firstPort {
		vportToAssign = firstPort
	} else {
		upperLimit := lastPort
		if assignedBiggest < lastPort {
			upperLimit = assignedBiggest
		}
		/*
		限定不指定业务的自动端口分配只能分配firstPort-lastPort区间的端口号
		*/
		for begin := firstPort; begin < upperLimit; begin++ {
			if portSlots[begin - 1000] == false {
				vportToAssign = begin
				break
			}
		}
		if vportToAssign == -1 {
			/*
			若未分配到端口，原因有二：
			1.firstPort-lastPort所有端口都已分配完
			2.可分配的端口大于已分配的最大端口号
			*/
			if upperLimit == lastPort {
				noAvailablePort = true
			}else {
				vportToAssign = assignedBiggest + 1
			}
		}
	}
	return
}

// 申请虚拟ip端口请求处理函数
func applyVPort(w http.ResponseWriter, r *http.Request) {

	autoOrNot, _ := strconv.Atoi(r.FormValue("autoornot"))
	business := -1
	var port int
	if autoOrNot == 1 && appConf.BusinessList != "" {
		business, _ = strconv.Atoi(r.FormValue("business"))
	}
	if autoOrNot == 0 {
		port, _ = strconv.Atoi(r.FormValue("port"))
	}
	servers := r.FormValue("servers")
	comment := strings.TrimSpace(r.FormValue("comment"))
	logOrNot := r.FormValue("logornot")

	rows, err := db.QueryVPort()
	if err != nil {
		logger.Println(err)
	}
	rowNum := len(rows)

	vportToAssign := -1

	dupOrNot := 0
	noAvailablePort := false

	// 若是指定端口方式，则判断指定的端口是否在1000-99999范围内
	inPortRange := true
	if autoOrNot == 0 && port < 1000 || port > 99999 {
		inPortRange = false
	}

	if inPortRange == true {
		/*
		将1000到已分配的最大端口号之间所有未占用和已占用的端口映射到一个真假值数组
		*/
		var portSlots []bool
		if rowNum > 0 {
			assignedBiggest := rows[rowNum - 1]
			allowedSmallest := 1000
			mayAssignedPortRange := assignedBiggest - allowedSmallest + 1
			//fmt.Println(mayAssignedPortRange)
			portSlots = make([]bool, mayAssignedPortRange)
			/*
			* 注意上一句make的用法-容量和长度都为mayAssignedPortRange,
			* 并且所有bool类型元素都自动初始化为false，所以下面的几行初始化代码不再需要。
			*/
			/*
			fmt.Println(portSlots[0])
			fmt.Println(portSlots[1])
			fmt.Println(portSlots[mayAssignedPortRange-1])
			for index := 0; index < mayAssignedPortRange; index++ {
				portSlots[index] = false
			}
			*/
			for index := 0; index < rowNum; index++ {
				port := rows[index]
				portSlots[port - 1000] = true
			}
		}
		// 自动分配端口
		if autoOrNot == 1 {
			// 未指定业务
			if business == -1 {
				/*
					虚拟ip端口自动分配算法(不指定业务)
					可分配端口范围：10000 - 19999
				*/
				if rowNum == 0 {
					vportToAssign = 10000
				}else {
					vportToAssign, noAvailablePort = autoAssignPort(10000, 19999, assignedBiggest, portSlots)
				}
			}else {
				// 指定业务
				var portRange string
				businesses := strings.Split(appConf.BusinessList, ";")
				for index, bToPortRange := range businesses {
					if index == business {
						thisBToPortRange := strings.Split(bToPortRange, ",")
						if comment != "" {
							comment += "<br />"
						}
						comment += "业务：" + thisBToPortRange[0]
						portRange = thisBToPortRange[1]
						break
					}
				}
				firstAndLast := strings.Split(portRange, "-")
				firstPort, _ := strconv.Atoi(firstAndLast[0])
				lastPort, _ := strconv.Atoi(firstAndLast[1])
				if rowNum == 0 {
					vportToAssign = firstPort
				}else {
					vportToAssign, noAvailablePort = autoAssignPort(firstPort, lastPort, assignedBiggest, portSlots)
				}
			}
		}else {
			// 指定端口
			// 检测端口是否已被占用
			for _, vport := range rows {
				if port == vport {
					dupOrNot = 1
					break
				}
			}
			if dupOrNot == 0 {
				vportToAssign = port
			}
		}
	}

	var result statusResult

	if inPortRange == false {
		result = statusResult{
			Success: "false",
			Msg: "指定的端口不在1000-99999的范围内！",
		}
	}else {
		if dupOrNot == 0 && noAvailablePort == false {
			now := time.Now().Format("2006-01-02 15:04:05")
			logornot, _ := strconv.Atoi(logOrNot)
			//fmt.Printf("servers: %s, vportToAssign: %d, comment: %s, logornot: %d, now: %s", servers, vportToAssign, comment, logornot, now)
			err = db.InsertNewTask(servers, vportToAssign, comment, logornot, now)
			if err != nil {
				logger.Println(err)
			}
			messageParts := make([]string,0, 2)
			messageParts = append(messageParts, appConf.Vip)
			messageParts = append(messageParts, strconv.Itoa(vportToAssign))
			message := strings.Join(messageParts, ":")

			result = statusResult{
				Success: "true",
				Msg:     message,
			}
		}else if (dupOrNot != 0) {
			result = statusResult{
				Success: "false",
				Msg: "端口已被占用，请选择指定其他端口！",
			}
		}else {
			result = statusResult{
				Success: "false",
				Msg: "该业务已没有可用的端口！",
			}
		}
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
		Comment  template.HTML
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

	var listenTasks = make([]listenTaskInfo,0, 100)
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
			Comment:  template.HTML(strings.Join(strings.Split(row.Comment, "\n"), "<br />")),
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
		bytes, err := ioutil.ReadFile(appConf.NewHAProxyConfPath)
		masterConf, err := os.OpenFile(appConf.MasterConf, os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0666)
		defer masterConf.Close()
		masterConf.Write(bytes)
		cmd := appConf.MasterRestartScript
		cmdToRun := exec.Command(cmd)
		err = cmdToRun.Run()
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
	appConf, err = config.ParseConfig("../conf/app_conf.ini")
	if err != nil {
		fmt.Println(err)
		return
	}

	port := flag.String("p", "9090", "port to run the web server")
	toolMode := flag.Bool("t", false, "run this program as a tool to export data from database to json or from json to database")

	flag.Parse()

	if *toolMode {
		// 数据转换存储方式
		err := tools.StorageTransform(appConf)
		tools.CheckError(err)
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
		err = http.ListenAndServe(":" + *port, nil)
		if err != nil {
			logger.Fatalln("ListenAndServe: ", err)
		}
	}
}
