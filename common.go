package main

import (
	"database/sql"
	"os"

	log "gopkg.in/inconshreveable/log15.v2"

	"time"

	"fmt"

	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/spf13/viper"
	"gopkg.in/gorp.v1"
)

type DbConfig struct {
	Hostname string
	Port     int64
	Username string
	Password string
	Database string
}

type Config struct {
	Db DbConfig
}

type UserResult struct {
	Mobile_number string  `db:"mobile_number"`
	Email         string  `db:"email"`
	Firstname     string  `db:"firstname"`
	Lastname      string  `db:"lastname"`
	Date_of_birth []uint8 `db:"date_of_birth"`
	Gender        string  `db:"gender"`
}

func InitDb() (dbmap *gorp.DbMap, config Config, err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if strings.Compare(os.Getenv("RUNMODE"), "testing") == 0 {
		viper.AddConfigPath("$HOME")
	}

	err = viper.ReadInConfig()
	ErrHandler(err)
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// log.Fataln("No config file found.")
			log.Crit("No config file found.")
			os.Exit(2)
			// Config file not found; ignore error if desired
		} else {
			// log.Fatalln(err)
			log.Crit(err.Error())
			os.Exit(2)
			// Config file was found but another error was produced
		}
	}

	config.Db.Hostname = viper.GetString("database.hostname")
	config.Db.Port = viper.GetInt64("database.port")
	config.Db.Username = viper.GetString("database.username")
	config.Db.Password = viper.GetString("database.password")
	config.Db.Database = viper.GetString("database.database")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Db.Username,
		config.Db.Password,
		config.Db.Hostname,
		config.Db.Port,
		config.Db.Database,
	))
	ErrHandler(err)
	if err != nil {
		return
	}

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	dbmap.AddTableWithName(User{}, "user").SetUniqueTogether("email", "mobile_number").SetKeys(false, "mobile_number")

	err = dbmap.CreateTablesIfNotExists()
	ErrHandler(err)
	if err != nil {
		return
	}

	return
}

func InsertUser(user UserIn) (err error) {
	Dbmap, Cfg, err = InitDb()
	ErrHandler(err)
	t, _ := time.Parse("02-01-2006", user.DateOfBirth)
	// ErrHandler(err)

	u := User{
		MobileNumber: user.Mobile,
		Email:        user.Email,
		Firstname:    user.Firstname,
		Lastname:     user.Lastname,
		DateOfBirth:  t,
	}

	// if strings.Compare(user.Gender, "") == 0 {
	// 	u.Gender = "Prefer not to mention"
	// }

	// log.Debug(fmt.Sprintf("%+v", "error debug"))
	err = Dbmap.Insert(&u)
	log.Debug("dbMap at common", "asdf", Dbmap)
	if err != nil {
		log.Debug(fmt.Sprintf("err %+v", err))
	}
	// log.Error(fmt.Sprintf("err %+v", err))
	// ErrHandler(err)

	return
}

func GetUser(user UserIn) (result UserIn, err error) {
	Dbmap, Cfg, err = InitDb()
	ErrHandler(err)

	var res UserResult
	var q = fmt.Sprintf("select * from user where mobile_number = '%s' and email = '%s'", user.Mobile, user.Email)
	log.Debug("Query of getUser", "q", q)

	err = Dbmap.SelectOne(&res, q)
	ErrHandler(err)

	log.Debug("query input", "user", user)
	log.Debug("result of database", "res", res)
	log.Debug("dob", "res", string(res.Date_of_birth))

	result.Mobile = res.Mobile_number
	result.Email = res.Email
	result.Firstname = res.Firstname
	result.Lastname = res.Lastname
	if strings.Compare(string(res.Date_of_birth), "0000-00-00 00:00:00") == 0 {
		res.Date_of_birth = []uint8{}
	}
	if (len(res.Date_of_birth) != 0) ||
		(strings.Compare(string(res.Date_of_birth), "") != 0) {
		t, err := time.Parse("2006-01-02 15:04:05", string(res.Date_of_birth))
		log.Debug("Content of res.Date_of_birth", "string(res.Date_of_birth)", string(res.Date_of_birth))
		log.Debug("Content of res.Date_of_birth", "res.Date_of_birth", res.Date_of_birth)
		log.Debug("Content of res.Date_of_birth", "len(res.Date_of_birth)", len(res.Date_of_birth))
		ErrHandler(err)
		result.DateOfBirth = t.Format("02-01-2006")
	}
	result.Gender = res.Gender

	return
}
