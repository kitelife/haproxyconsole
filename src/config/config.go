package config

import (
	"errors"
	//"fmt"
	"github.com/robfig/config"
	"strconv"
	"strings"
	"os"
	"path/filepath"
	"regexp"
)

type ConfigInfo struct {
	BusinessList		string
	MasterConf          string
	MasterRestartScript string
	SlaveServer         string
	SlaveRemoteUser     string
	SlaveRemotePasswd   string
	SlaveConf           string
	SlaveRestartScript  string
	StoreScheme         int
	DBDriverName        string
	DBDataSourceName    string
	FileToReplaceDB     string
	MasterStatsPage     string
	SlaveStatsPage      string
	Vip                 string
	NewHAProxyConfPath  string
}

// 函数checkBusinessList使用
func quickSort(portRangeList [][2]int, left int, right int) {
	temp := portRangeList[left]
	p := left
	i, j := left, right

	for i <= j {
		for j >= p && portRangeList[j][0] >= temp[0] {
			j--
		}
		if j >= p {
			portRangeList[p] = portRangeList[j]
			p = j
		}

		if portRangeList[i][0] <= temp[0] && i <= p {
			i++
		}

		if i <= p {
			portRangeList[p] = portRangeList[i]
			p = i
		}
	}
	portRangeList[p] = temp
	if p - left > 1 {
		quickSort(portRangeList, left, p - 1)
	}
	if right - p > 1 {
		quickSort(portRangeList, p + 1, right)
	}
}

// 检查业务列表配置项是否正确
func checkBusinessList(bl string) (err error) {
	// 允许使用1000-99999范围内的端口
	matched, _ := regexp.MatchString(`^((.+,\d{4,5}-\d{4,5};)*(.+,\d{4,5}-\d{4,5}))?$`, bl)
	if matched == false {
		err = errors.New("启动失败：业务端口区间列表BusinessList配置的值有误！请检查！")
		return
	}
	// 检查端口范围的开始值是否大于结束值，以及业务端口范围是否有重叠
	// 预估业务类型数目不超过15个
	businesses := strings.Split(bl, ";")
	nameToPortRanges := make(map[string][2]int)
	portRangeList := make([][2]int,0, 15)
	for _, business := range businesses {
		nameToPortRange := strings.Split(business, ",")
		portRange := strings.Split(nameToPortRange[1], "-")
		beginPort, _ := strconv.Atoi(portRange[0])
		endPort, _ := strconv.Atoi(portRange[1])
		nameToPortRanges[nameToPortRange[0]] = [2]int{ beginPort, endPort}
		portRangeList = append(portRangeList, [2]int{beginPort, endPort})
	}

	// 端口范围的开始值是否大于结束值
	beginGtEnd := false
	beginGtEndBusiness := make([]string,0, 15)
	for businessName, portRange := range nameToPortRanges {
		if portRange[0] > portRange[1] || portRange[0] < 1000 || portRange[1] > 99999 {
			beginGtEnd = true
			beginGtEndBusiness = append(beginGtEndBusiness, businessName)
		}
	}
	if beginGtEnd {
		err = errors.New("启动失败：" + strings.Join(beginGtEndBusiness, ",") + "的端口范围有误！")
		return
	}

	// 业务端口范围是否有重叠
	isOverlap := false
	// 先对业务端口范围按照范围的起始端口从小到大排序
	rangeNum := len(portRangeList)
	quickSort(portRangeList, 0, rangeNum - 1)
	for index := 0; index < rangeNum - 1; index++ {
		if portRangeList[index][1] >= portRangeList[index + 1][0] {
			isOverlap = true
			break
		}
	}
	if isOverlap {
		err = errors.New("启动失败：配置文件中BusinessList配置的业务端口区间有重叠！")
		return
	}
	return
}

// 检查配置文件中[master]部分配置的正确性
func checkMaster(mc string, mrs string) (err error) {
	// 检查MasterConf指定的主HAProxy配置文件是否存在
	if _, e := os.Stat(mc); os.IsNotExist(e) {
		err = errors.New("启动失败：配置文件[master]部分中MasterConf指定的主HAProxy配置文件不存在！")
		return
	}

	// 检查MasterRestartScript指定的主HAProxy重启脚本是否存在
	if _, e := os.Stat(mrs); os.IsNotExist(e) {
		err = errors.New("启动失败：配置文件中[master]部分中MasterRestartScript指定的主HAProxy重启脚本不存在！")
		return
	}
	return
}

// 检查配置文件中[store]部分配置的正确性
func checkStore(conf ConfigInfo) (err error) {
	// 采用json文件存储时
	if conf.StoreScheme == 1 {
		storeDir := filepath.Dir(conf.FileToReplaceDB)
		if _, err = os.Stat(storeDir); os.IsNotExist(err) || filepath.Ext(conf.FileToReplaceDB) != ".json" {
			err = errors.New("启动失败：配置文件中[store]部分的FileToReplaceDB项配置有误！")
			return
		}
	}else if conf.StoreScheme == 0 { // 采用数据库存储时
		// DSN(Data Source Name)的格式：[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
		matched, _ := regexp.MatchString(`^.+:.*@(tcp(4|6)?|udp(4|6)?|ip(4|6)?|unix(gram|packet)?)\(.+\)/.+(\?.+=.+(&.+=.+)*)?$`, conf.DBDataSourceName)
		if conf.DBDriverName != "mysql" || matched == false {
			err = errors.New("启动失败：配置文件[store]部分的DBDriverName或DBDataSourceName配置有误！")
			return
		}
	}else {    //配置项StoreScheme有误
		err = errors.New("启动失败：配置文件[store]部分的StoreScheme项配置有误")
		return
	}
	return
}

// 检查配置文件[stats]部分配置的正确性
func checkStats(msp string, ssp string) (err error) {
	// \d{2,6}中,和6之间不能有空格
	pattern := regexp.MustCompile(`^http://.+:\d{2,6}$`)
	mspMatched := pattern.MatchString(msp)
	sspMatched := pattern.MatchString(ssp)
	if mspMatched == false || sspMatched == false {
		err = errors.New("启动失败：配置文件[stats]部分的配置项有误！")
		return
	}
	return
}

// 检查配置文件中配置项的正确性
func CheckConfig(conf ConfigInfo) (err error) {
/*
	err = checkBusinessList(conf.BusinessList)
	if err != nil {
		return
	}
*/
	err = checkMaster(conf.MasterConf, conf.MasterRestartScript)
	if err != nil {
		return
	}

	err = checkStore(conf)
	if err != nil {
		return
	}

	err = checkStats(conf.MasterStatsPage, conf.SlaveStatsPage)
	if err != nil {
		return
	}
	return
}

func ParseConfig(configPath string) (ci ConfigInfo, err error) {
	conf, err := config.ReadDefault(configPath)
	if err != nil {
		return
	}
	businessList, _ := conf.String("mode", "BusinessList")
	masterConf, _ := conf.String("master", "MasterConf")
	masterRestartScript, _ := conf.String("master", "MasterRestartScript")

	slaveServer, _ := conf.String("slave", "SlaveServer")
	slaveRemoteUser, _ := conf.String("slave", "SlaveRemoteUser")
	slaveRemotePasswd, _ := conf.String("slave", "SlaveRemotePasswd")

	slaveConf, _ := conf.String("slave", "SlaveConf")
	slaveRestartScript, _ := conf.String("slave", "SlaveRestartScript")

	storeScheme, _ := conf.Int("store", "StoreScheme")

	dbDriverName, _ := conf.String("store", "DBDriverName")
	dbDataSourceName, _ := conf.String("store", "DBDataSourceName")

	fileToReplaceDB, _ := conf.String("store", "FileToReplaceDB")

	masterStatsPage, _ := conf.String("stats", "MasterStatsPage")
	slaveStatsPage, _ := conf.String("stats", "SlaveStatsPage")

	vip, _ := conf.String("others", "Vip")

	newHAProxyConfPath, _ := conf.String("others", "NewHAProxyConfPath")

	ci = ConfigInfo{
		BusinessList:         businessList,
		MasterConf:          masterConf,
		MasterRestartScript: masterRestartScript,
		SlaveServer:         slaveServer,
		SlaveRemoteUser:     slaveRemoteUser,
		SlaveRemotePasswd:   slaveRemotePasswd,
		SlaveConf:           slaveConf,
		SlaveRestartScript:  slaveRestartScript,
		StoreScheme:         storeScheme,
		DBDriverName:        dbDriverName,
		DBDataSourceName:    dbDataSourceName,
		FileToReplaceDB:     fileToReplaceDB,
		MasterStatsPage:     masterStatsPage,
		SlaveStatsPage:      slaveStatsPage,
		Vip:                 vip,
		NewHAProxyConfPath:  newHAProxyConfPath,
	}
	err = CheckConfig(ci)
	return
}