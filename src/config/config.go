package config

import (
	"errors"
	"github.com/robfig/config"
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

// 检查业务列表配置项是否正确
func checkBusinessList(bl string) (err error) {
	// 允许使用1000-100000范围内的端口
	matched, _ := regexp.MatchString(`^((.+,\d{4,6}-\d{4,6};)*(.+,\d{4,6}-\d{4,6}))?$`, bl)
	if matched == false {
		err = errors.New("启动失败：业务端口区间列表BusinessList配置的值有误！请检查！")
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
func checkStats(msp string, ssp string)(err error){
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
	err = checkBusinessList(conf.BusinessList)
	if err != nil {
		return
	}

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
