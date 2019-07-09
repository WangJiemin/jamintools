package mydb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/juju/errors"

	_ "github.com/go-sql-driver/mysql"
	"github.com/WangJiemin/jamintools/ehand"
)

const (
	C_mysql_alive_query                 string = "select 1"
	C_mysql_sql_global_status           string = "show global status"
	C_mysql_sql_global_vars             string = "show global variables"
	C_mysql_sql_slave_status            string = "show slave status"
	C_mysql_sql_slave_status_all        string = "show all slaves status"
	C_mysql_sql_master_status           string = "show master status"
	C_mysql_sql_innodb_status           string = "show engine innodb status"
	C_mysql_sql_disable_readonly        string = "set global read_only = 0"
	C_mysql_sql_enable_readonly         string = "set global read_only = 1"
	C_mysql_sql_disable_super_readonly  string = "set global super_read_only=0"
	C_mysql_sql_enable_super_readonly   string = "set global super_read_only=1"
	C_mysql_sql_disable_event_scheduler string = "set global event_scheduler=0"
	C_mysql_sql_enable_event_scheduler  string = "set global event_scheduler=1"
	C_mysql_sql_unlock_all_tables       string = "unlock tables"
	C_mysql_sql_var_read_only           string = "select @@global.read_only as val"
	C_mysql_sql_var_super_read_only     string = "select @@global.super_read_only as val"
	C_mysql_sql_var_event_scheduler     string = "select @@global.event_scheduler as val"
)

func UnlockAllTables(db *sql.DB) error {
	_, err := db.Exec(C_mysql_sql_unlock_all_tables)
	return err
}

func EnableEventScheduler(db *sql.DB) error {
	result, err := SetAndQueryGlobalVars(db, "event_scheduler", "1")
	if result == "" {
		return fmt.Errorf("error to enable event_scheduler: %s", err)
	}
	if err != nil {
		return fmt.Errorf("OK to enable event_scheduler, but fail to read back value of event_scheduler: %s", err)
	}
	if result != "ON" {
		return fmt.Errorf("OK to enable event_scheduler, but then read back, the value of event_scheduler is %s, not expected %s", result, "ON")
	}
	return nil
}

func DisableEventScheduler(db *sql.DB) error {
	result, err := SetAndQueryGlobalVars(db, "event_scheduler", "0")
	if result == "" {
		return fmt.Errorf("error to disable event_scheduler: %s", err)
	}
	if err != nil {
		return fmt.Errorf("OK to disable event_scheduler, but fail to read back value of event_scheduler: %s", err)
	}
	if result != "OFF" {
		return fmt.Errorf("OK to disable event_scheduler, but then read back, the value of event_scheduler is %s, not expected %s", result, "OFF")
	}
	return nil
}

func EnableSuperReadOnly(db *sql.DB) error {
	result, err := SetAndQueryGlobalVars(db, "super_read_only", "1")
	if result == "" {
		return fmt.Errorf("error to enable super_read_only: %s", err)
	}
	if err != nil {
		return fmt.Errorf("OK to enable super_read_only, but fail to read back value of super_read_only: %s", err)
	}
	if result != "ON" {
		return fmt.Errorf("OK to enable super_read_only, but then read back, the value of super_read_only is %s, not expected %s", result, "ON")
	}
	return nil
}

func DisableSuperReadOnly(db *sql.DB) error {
	result, err := SetAndQueryGlobalVars(db, "super_read_only", "0")
	if result == "" {
		return fmt.Errorf("error to disable super_read_only: %s", err)
	}
	if err != nil {
		return fmt.Errorf("OK to disable super_read_only, but fail to read back value of super_read_only: %s", err)
	}
	if result != "OFF" {
		return fmt.Errorf("OK to disable super_read_only, but then read back, the value of super_read_only is %s, not expected %s", result, "OFF")
	}
	return nil
}

func EnableReadOnly(db *sql.DB) error {
	result, err := SetAndQueryGlobalVars(db, "read_only", "1")
	if result == "" {
		return fmt.Errorf("error to enable read_only: %s", err)
	}
	if err != nil {
		return fmt.Errorf("OK to enable read_only, but fail to read back value of read_only: %s", err)
	}
	if result != "1" {
		return fmt.Errorf("OK to enable read_only, but then read back, the value of read_only is %s, not expected %s", result, "1")
	}
	return nil
}

func DisableReadOnly(db *sql.DB) error {
	result, err := SetAndQueryGlobalVars(db, "read_only", "0")
	if result == "" {
		return fmt.Errorf("error to disable read_only: %s", err)
	}
	if err != nil {
		return fmt.Errorf("OK to disable read_only, but fail to read back value of read_only: %s", err)
	}
	if result != "0" {
		return fmt.Errorf("OK to diable read_only, but then read back, the value of read_only is %s, not expected %s", result, "0")
	}
	return nil
}

// return NotFound error if Unknown system variable
func GetVariableValueStr(db *sql.DB, varName string, ifGlobal bool) (string, error) {
	var (
		queryStr string
		err      error
		val      string
	)
	if ifGlobal {
		queryStr = fmt.Sprintf("select @@global.%s as val", varName)
	} else {
		queryStr = fmt.Sprintf("select @@session.%s as val", varName)
	}
	row := db.QueryRow(queryStr)
	err = row.Scan(&val)
	if err != nil {
		if strings.Contains(err.Error(), "Unknown system variable") {
			return "", errors.NewNotFound(err, "no such variable "+varName)
		}
		return "", errors.Annotatef(err, "error to get value of var %s", varName)
	} else {
		return val, nil
	}
}

// return varValue
func SetAndQueryGlobalVars(db *sql.DB, varName string, varValue string) (string, error) {
	setSql := fmt.Sprintf("set global %s = %s", varName, varValue)
	_, err := db.Exec(setSql)
	if err != nil {
		return "", err
	}
	querySql := fmt.Sprintf("select @@global.%s as val", varName)
	rows, err := db.Query(querySql)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return varValue, err
	}
	var (
		setVal string
	)
	for rows.Next() {
		err = rows.Scan(&setVal)
		if err != nil {
			return varValue, err
		}
		return setVal, nil
	}
	return setVal, nil
}

func GetMyConnectionId(db *sql.DB) (int, error) {
	var (
		id int
	)
	rows, err := db.Query("select connection_id() as id")
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return -1, err
	}
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return -1, err
		} else {
			return id, nil
		}
	}
	return id, nil
}

func MergeMapUint64(dst map[string]uint64, src map[string]uint64) map[string]uint64 {
	mg := map[string]uint64{}
	for k, v := range src {
		mg[k] = v
	}
	for k, v := range dst {
		mg[k] = v
	}
	return mg
}

func CheckMysqlAlive(db *sql.DB, querySql string) bool {
	rows, err := db.Query(querySql)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		if CheckIfMysqlAliveError(err) {
			return true
		} else {
			return false
		}
	} else {

		return true
	}
}

func MysqlShowGlobalStatus(db *sql.DB) (map[string]int64, error) {
	var (
		err      error
		sts      map[string]int64 = map[string]int64{}
		varName  string
		varValue int64
	)
	rows, err := db.Query(C_mysql_sql_global_status)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return sts, ehand.WithStackError(err)
	}

	for rows.Next() {
		err = rows.Scan(&varName, &varValue)
		if err != nil {
			//ignore error, example string value
			continue
		}
		sts[varName] = varValue
	}

	if len(sts) == 0 {
		return sts, ehand.WithStackError(err)
	}
	return sts, nil

}

func CheckBinlogFormatRowFull(db *sql.DB) error {
	val, err := GetVariableValueStr(db, "binlog_format", true)
	if err != nil {
		return err
	}
	if val != "ROW" {
		return errors.Errorf("binlog_format=%s, must be ROW", val)
	}
	val, err = GetVariableValueStr(db, "binlog_row_image", true)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	} else if val != "FULL" {
		return errors.Errorf("binlog_row_image=%s, must be FULL", val)
	} else {
		return nil
	}
}
