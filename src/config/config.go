package config

import (
	"errors"
	"regexp"
	"github.com/robfig/config"
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

// 检查配置文件中配置项的正确性
func CheckConfig(conf ConfigInfo) (err error) {
	err = checkBusinessList(conf.BusinessList)
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
