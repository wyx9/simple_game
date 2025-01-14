package pkg

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"simple_game/game/config"
)

var DB *sqlx.DB

func MysqlStart() {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Conf.MySql.User, config.Conf.MySql.PassWord, config.Conf.MySql.Addr, config.Conf.MySql.Port, config.Conf.MySql.DBName)
	database, err := sqlx.Open("mysql", dataSourceName)
	if err != nil || database == nil {
		fmt.Println("mysql conn - fail-", err)
		return
	} else {
		INFO("mysql server start suc")
	}
	DB = database
}
