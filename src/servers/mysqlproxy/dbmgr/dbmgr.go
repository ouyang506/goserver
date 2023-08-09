package dbmgr

import (
	"database/sql"
	"fmt"
	"framework/log"
	"mysqlproxy/config"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MysqlCharset = "utf8"
)

type MysqlMgr struct {
	db *sql.DB
}

var (
	once               = sync.Once{}
	mysqlMgr *MysqlMgr = nil
)

// singleton
func GetMysqlMgr() *MysqlMgr {
	once.Do(func() {
		mysqlMgr = newMysqlMgr()
	})
	return mysqlMgr
}

func newMysqlMgr() *MysqlMgr {
	mgr := &MysqlMgr{}
	return mgr
}

func (mgr *MysqlMgr) Start() {
	mgr.doConnect()
}

func (mgr *MysqlMgr) doConnect() {
	conf := config.GetConfig()
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		conf.MysqlConf.Username, conf.MysqlConf.Password,
		conf.MysqlConf.IP, conf.MysqlConf.Port, conf.MysqlConf.Database, MysqlCharset)
	var err error = nil
	mgr.db, err = sql.Open("mysql", dataSource)
	if err != nil {
		log.Error("open mysql driver error: %s", err)
		return
	}
	mgr.db.SetMaxOpenConns(conf.MysqlConf.PoolMaxConn)
	mgr.db.SetMaxIdleConns(int(conf.MysqlConf.PoolMaxConn / 2))
}

func (mgr *MysqlMgr) QuerySql(sql string, args ...any) (columns []string, result [][]string, err error) {
	stmt, err := mgr.db.Prepare(sql)
	if err != nil {
		log.Error("prepare statement error: %v", err)
		return
	}
	defer stmt.Close()

	row, err := stmt.Query(args...)
	if err != nil {
		log.Error("query sql error : %v, sql : %s", err, sql)
		return
	}
	defer row.Close()

	columns, err = row.Columns()
	if err != nil {
		log.Error("get columns error : %v", err)
		return
	}

	dst := make([]any, len(columns))
	dstptr := make([]any, len(columns))
	for i := range dstptr {
		dstptr[i] = &dst[i]
	}
	for row.Next() {
		err = row.Scan(dstptr...)
		if err == nil {
			log.Error("scan error: %v", err)
			return
		}

		rowStrArr := make([]string, len(dst))
		for i, v := range dst {
			valueStr := ""
			switch tmp := v.(type) {
			case string:
				valueStr = tmp
			case int, int8, int16, int32, int64, uint, uint8, uint32, uint64:
				valueStr = fmt.Sprintf("%v", v)
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
			rowStrArr[i] = valueStr
		}

		result = append(result, rowStrArr)
	}

	return
}

func (mgr *MysqlMgr) ExecuteSql(sql string, args ...any) (lastInsertId int64, rowsAffected int64, err error) {
	stmt, err := mgr.db.Prepare(sql)
	if err != nil {
		log.Error("prepare statement error: %v", err)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		log.Error("execute sql error: %v, sql: %v", err, sql)
		return
	}

	lastInsertId, err = result.LastInsertId()
	if err != nil {
		log.Error("get last insert id error: %v", err)
		return
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		log.Error("get rows affected error: %v", err)
		return
	}
	return
}
