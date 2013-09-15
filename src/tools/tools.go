package tools

import (
	"config"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const sqlToInit = `DROP TABLE IF EXISTS haproxymapinfo;
CREATE TABLE haproxymapinfo (
	id int(11) NOT NULL AUTO_INCREMENT,
 	servers varchar(1024) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
 	vport int(10) NOT NULL,
 	comment varchar(1024) DEFAULT '',
 	logornot int(1) DEFAULT '1',
 	datetime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 	PRIMARY KEY (id)
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func InitDataTable(appConf config.ConfigInfo) {
	db, err := sql.Open(appConf.DBDriverName, appConf.DBDataSourceName)
	defer db.Close()
	CheckError(err)
	_, err = db.Exec(sqlToInit)
	CheckError(err)
	return
}

func StorageTransform(appConf config.ConfigInfo) (err error) {

	type DataRow struct {
		Id       int
		Servers  string
		VPort    int
		Comment  string
		LogOrNot int
		DateTime string
	}

	db, err := sql.Open(appConf.DBDriverName, appConf.DBDataSourceName)
	defer db.Close()
	CheckError(err)

	if appConf.StoreScheme == 0 {
		rows, err := db.Query("SELECT id, servers, vport, comment, logornot, datetime FROM haproxymapinfo ORDER BY vport ASC")
		CheckError(err)
		var id int
		var servers string
		var vport int
		var comment string
		var logornot int
		var datetime string
		taskList := make([]DataRow, 0, 100)
		for rows.Next() {
			err = rows.Scan(&id, &servers, &vport, &comment, &logornot, &datetime)
			taskList = append(taskList, DataRow{Id: id, Servers: servers, VPort: vport, Comment: comment, LogOrNot: logornot, DateTime: datetime})
		}
		dataJson, err := json.MarshalIndent(taskList, "", "    ")
		CheckError(err)
		f, err := os.OpenFile(appConf.FileToReplaceDB, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		CheckError(err)
		defer f.Close()
		f.Write(dataJson)
		f.Sync()
		return nil
	}
	if appConf.StoreScheme == 1 {
		bytes, err := ioutil.ReadFile(appConf.FileToReplaceDB)
		allData := make([]DataRow, 0, 100)
		err = json.Unmarshal(bytes, &allData)
		CheckError(err)
		// 这里还得先测试数据表haproxy是否存在，若不存在，则需创建
		_, err = db.Exec(sqlToInit)
		CheckError(err)
		result, err := db.Exec("INSERT INTO haproxymapinfo (id, servers, vport, comment, logornot, datetime) VALUES (?, ?, ?, ?, ?, ?)", allData)
		CheckError(err)
		num, err := result.RowsAffected()
		CheckError(err)
		fmt.Println(num)
		return nil
	}

	return errors.New("存储方式不正确")
}
