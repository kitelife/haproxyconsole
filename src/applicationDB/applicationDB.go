package applicationDB

import (
	"bytes"
	"config"
	"database/sql"
	"encoding/json"
	"io"
	"os"
)

type NewConfDataType struct {
	Servers  string
	VPort    int
	LogOrNot int
}

type DataRow struct {
	Id       int
	Servers  string
	VPort    int
	Comment  string
	LogOrNot int
	DateTime string
}

type DB interface {
	QueryNewConfData() ([]NewConfDataType, error)
	QueryVPort() ([]int, error)
	InsertNewTask(string, int, string, int, string) (error)
	QueryForTaskList() ([]DataRow, error)
	DeleteTask(int) (sql.Result, error)
	UpdateTaskInfo(string, string, int, string, int) (error)
}

type database sql.DB
type fileForStore os.File

// 该辅助函数来自golang标准库io/ioutil/ioutil.go
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte,0, capacity))

	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

func readJson(f fileForStore) (allData []DataRow, err error) {
	var n int64

	if fi, err := f.Stat(); err == nil {
		if size := fi.Size(); size < 1e9 {
			n = size
		}
	}
	content, err := readAll(f, n + bytes.MinRead)
	err = json.Unmarshal(content, &allData)
	return
}

// 模拟数据库增删改操作的返回结果Result
/*
type Result interface {
        LastInsertId() (int64, error)
        RowsAffected() (int64, error)
}
*/
type ResultDeleteFromFile struct {
	LastIdInsert int64
	AffectedRows int64
}

func (rdff ResultDeleteFromFile) LastInsertId() (int64, error) {
	return rdff.LastIdInsert, nil
}

func (rdff ResultDeleteFromFile) RowsAffected() (int64, error) {
	return rdff.AffectedRows, nil
}

// SELECT servers, vport, logornot FROM haproxymapinfo ORDER BY vport ASC
// QueryNewConfData() ([]NewConfDataType, error)
func (d database) QueryNewConfData() (dataList []NewConfDataType, err error) {
	dataList = make([]NewConfDataType,0, 100)
	rows, err := d.Query("SELECT servers, vport, logornot FROM haproxymapinfo ORDER BY vport ASC")
	var servers string
	var vport int
	var logOrNot int
	for rows.Next() {
		err = rows.Scan(&servers, &vport, &logOrNot)
		dataList = append(dataList, NewConfDataType{Servers: servers, VPort: vport, LogOrNot: logOrNot})
	}
	return
}

func (f fileForStore) QueryNewConfData() (dataList []NewConfDataType, err error) {
	allData, err := readJson(f)
	dataList = make([]NewConfDataType,0, 100)
	taskNum := len(allData)
	for index := 0; index < taskNum; index++ {
		task := allData[index]
		dataList = append(dataList, NewConfDataType{Servers: task.Servers, VPort: task.VPort, LogOrNot: task.LogOrNot})
	}
	return
}

// SELECT vport FROM haproxymapinfo ORDER BY vport ASC
// QueryVPort() ([]int, error)
func (d database) QueryVPort() (vportList []int, err error) {
	rows, err := d.Query("SELECT vport FROM haproxymapinfo ORDER BY vport ASC")
	vportList = make([]int,0, 100)
	var vport int
	for rows.Next() {
		err = rows.Scan(&vport)
		vportList = append(vportList, vport)
	}
	return
}

func (f fileForStore) QueryVPort() (vportList []int, err error) {
	allData, err := readJson(f)
	vportList = make([]int, 100)
	taskNum := len(allData)
	for index := 0; index < taskNum; index++ {
		vportList = append(vportList, allData[index].VPort)
	}
	return
}

// db.Exec("INSERT INTO haproxymapinfo (servers, vport, comment, logornot, datetime) VALUES (?, ?, ?, ?, ?)", servers, vportToAssign, comment, logOrNot, now)
// InsertNewTask(string, int, string, int, string) (error)
func (d database) InsertNewTask(servers string, vportToAssign int , comment string, logOrNot int, now string) (err error) {
	_, err := d.Exec("INSERT INTO haproxymapinfo (servers, vport, comment, logornot, datetime) VALUES (?, ?, ?, ?, ?)", servers, vportToAssign, comment, logOrNot, now)
	return
}

func (f fileForStore) InsertNewTask(servers string, vportToAssign int, comment string, logOrNot int, now string) (err error) {
	allData, err := readJson(f)
	rowNum := len(allData)
	maxId := 0
	for index := 0; index < rowNum; index ++ {
		row := allData[index]
		if row.Id > maxId {
			maxId = row.Id
		}
	}
	oneRowData := DataRow{
		Id: maxId + 1,
		Servers: servers,
		VPort: vportToAssign,
		Comment: comment,
		LogOrNot: logOrNot,
		DateTime: now,
	}
	allData = append(allData, oneRowData)
	// 直接截断/清空文件，然后整个写入
	err = f.Truncate(0)
	dataJson, err := json.Marshal(allData)
	f.Write(dataJson)
	return
}

// SELECT id, servers, vport, comment, logornot, datetime FROM haproxymapinfo ORDER BY vport ASC
// QueryForTaskList() ([]DataRow, error)
func (d database) QueryForTaskList() (taskList []DataRow, err error) {
	rows, err := d.Query("SELECT id, servers, vport, comment, logornot, datetime FROM haproxymapinfo ORDER BY vport ASC")
	var id int
	var servers string
	var vport int
	var comment string
	var logornot int
	var datetime string
	taskList = make([]DataRow,0, 100)
	for rows.Next() {
		err = rows.Scan(&id, &servers, &vport, &comment, &logornot, &datetime)
		taskList = append(taskList, DataRow{Id: id, Servers: servers, VPort: vport, Comment: comment, LogOrNot: logornot, DateTime: datetime})
	}
	return
}

func (f fileForStore) QueryForTaskList() (taskList []DataRow, err error) {
	taskList, err := readJson(f)
	return
}

// db.Exec("DELETE FROM haproxymapinfo WHERE id=?", id)
// DeleteTask(int) (Result, error)
func (d database) DeleteTask(id int) (result sql.Result, err error) {
	result, err := d.Exec("DELETE FROM haproxymapinfo WHERE id=?", id)
	return
}

func (f fileForStore) DeleteTask(id int) (result sql.Result, err error) {
	allData, err := readJson(f)
	rowNum := len(allData)
	dataAfterDel := make([]DataRow,0, 100)
	for index := 0; index < rowNum; index++ {
		row := allData[index]
		if row.Id != id {
			dataAfterDel = append(dataAfterDel, row)
		}
	}
	err = f.Truncate(0)
	dataJson, err := json.Marshal(dataAfterDel)
	f.Write(dataJson)
	rdff := ResultDeleteFromFile{LastIdInsert:-1, AffectedRows: 1}
	return rdff, nil
}

// db.Exec("UPDATE haproxymapinfo SET servers=?, comment=?, logornot=?, datetime=? WHERE id=?", servers, comment, logornot, now, id)
// UpdateTaskInfo(string, string, int, string, int) (error)
func (d database) UpdateTaskInfo(servers string, comment string, logornot int, now string, id int) (err error) {
	_, err := d.Exec("UPDATE haproxymapinfo SET servers=?, comment=?, logornot=?, datetime=? WHERE id=?", servers, comment, logornot, now, id)
	return
}

func (f fileForStore) UpdateTaskInfo(servers string, comment string, logornot int, now string, id int) (err error) {
	allData, err := readJson(f)
	rowNum := len(allData)
	for index := 0; index < rowNum; index++ {
		row := allData[index]
		if row.Id == id {
			dataOneRow := DataRow{
				Id: id,
				Servers: servers,
				VPort: row.VPort,
				Comment: comment,
				LogOrNot: logornot,
				DateTime: now,
			}
			allData[index] = dataOneRow
		}
	}
	err = f.Truncate(0)
	dataJson, err := json.Marshal(allData)
	f.Write(dataJson)
	return
}

func InitStoreConnection(appConf config.ConfigInfo) (db DB, err error) {
	if appConf.StoreScheme == 0 {
		db, err = sql.Open(appConf.DBDriverName, appConf.DBDataSourceName)
	}

	if appConf.StoreScheme == 1 {
		db, err = os.OpenFile(appConf.FileToReplaceDB, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
	}
	return
}
