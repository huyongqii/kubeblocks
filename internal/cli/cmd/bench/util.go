/*
Copyright ApeCloud, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bench

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	unknownDB   = "Unknown database"
	createDBDDL = "CREATE DATABASE IF NOT EXISTS "
	mysqlDriver = "mysql"
)

func openDB() error {
	var (
		err error
		ds  = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbName)
	)

	// allow multiple statements in one query to allow q15 on the TPC-H
	fullDsn := fmt.Sprintf("%s?multiStatements=true", ds)
	globalDB, err = sql.Open(mysqlDriver, fullDsn)
	if err != nil {
		return err
	}

	return ping()
}

func ping() error {
	if globalDB == nil {
		return nil
	}
	if err := globalDB.Ping(); err != nil {
		errString := err.Error()
		if strings.Contains(errString, unknownDB) {
			return createDB()
		} else {
			globalDB = nil
		}
		return err
	}
	return nil
}

func createDB() error {
	tmpDs := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)
	tmpDB, _ := sql.Open(mysqlDriver, tmpDs)
	defer tmpDB.Close()
	if _, err := tmpDB.Exec(createDBDDL + dbName); err != nil {
		return fmt.Errorf("failed to create database, err %v", err)
	}
	return nil
}

func closeDB() error {
	if globalDB == nil {
		return nil
	}
	return globalDB.Close()
}
