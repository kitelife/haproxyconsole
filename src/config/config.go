package config

import "github.com/robfig/config"

type ConfigInfo struct {
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

func ParseConfig(configPath string) (ci ConfigInfo, err error) {
	conf, err := config.ReadDefault(configPath)

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

	return
}
