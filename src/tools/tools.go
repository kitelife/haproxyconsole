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

func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func StorageTransform(appConf config.ConfigInfo) (err error) {

	type DataRow struct {
		Id       int
		Servers  string
		BackupServers  string
		VPort    int
		Comment  string
		LogOrNot int
		DateTime string
	}

	db, err := sql.Open(appConf.DBDriverName, appConf.DBDataSourceName)
	defer db.Close()
	CheckError(err)

	if appConf.StoreScheme == 0 {
		fmt.Println("**从数据库读取数据存入JSON文件中**")
		rows, err := db.Query("SELECT id, servers, backup_servers, vport, comment, logornot, datetime FROM haproxymapinfo ORDER BY vport ASC")
		CheckError(err)
		var id int
		var servers string
		var backupServers string
		var vport int
		var comment string
		var logornot int
		var datetime string
		taskList := make([]DataRow, 0, 100)
		for rows.Next() {
			err = rows.Scan(&id, &servers, &backupServers, &vport, &comment, &logornot, &datetime)
			taskList = append(taskList, DataRow{Id: id, Servers: servers, BackupServers: backupServers, VPort: vport, Comment: comment, LogOrNot: logornot, DateTime: datetime})
		}
		fmt.Printf("共%d条数据\n", len(taskList))
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
		fmt.Println("**从JSON文件读取数据插入数据库中**")
		bytes, err := ioutil.ReadFile(appConf.FileToReplaceDB)
		allData := make([]DataRow, 0, 100)
		err = json.Unmarshal(bytes, &allData)
		CheckError(err)
		// 这里还得先测试数据表haproxy是否存在，若不存在，则需创建
		fmt.Printf("将插入%d条数据\n", len(allData))
		var num int64
		num = 0
		for _, data := range allData {
			result, err := db.Exec("INSERT INTO haproxymapinfo (id, servers, backup_servers, vport, comment, logornot, datetime) VALUES (?, ?, ?, ?, ?, ?)", data.Id, data.Servers, data.VPort, data.Comment, data.LogOrNot, data.DateTime)
			CheckError(err)
			n, err := result.RowsAffected()
			CheckError(err)
			num += n
		}
		fmt.Printf("共插入数据%d条\n", num)
		return nil
	}

	return errors.New("存储方式不正确")
}
