package model

import (
	"database/sql"
	"fmt"
	"github.com/Unknwon/goconfig"
	_ "github.com/go-sql-driver/mysql"
	//"fmt"
	//"github.com/Unknwon/goconfig"
)

type DbLinker struct {
	Dsn 	string
	DB  	*sql.DB
}

type QueryBuilder struct {
	Table	string
	Where	map[string]string
	Target	string
}
func (d *DbLinker) Init() {
	configPath 				:= "./config.ini"
	config, load_conf_err 	:= goconfig.LoadConfigFile(configPath)
	if load_conf_err 		!= nil {
		fmt.Println(load_conf_err)
	}

	dsn, _ 	:= config.GetValue("mysql", "dsn");
	db, _ 	:= sql.Open("mysql", dsn)

	err := db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	d.DB = db
}
