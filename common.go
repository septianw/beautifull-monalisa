package main

import (
	"database/sql"

	"time"

	"fmt"

	"errors"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	// "github.com/spf13/viper"
	"gopkg.in/gorp.v1"
)

// Store Database config fetched from config here
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

// Query result type holder
type UserResult struct {
	Mobile_number string  `db:"mobile_number"`
	Email         string  `db:"email"`
	Firstname     string  `db:"firstname"`
	Lastname      string  `db:"lastname"`
	Date_of_birth []uint8 `db:"date_of_birth"`
	Gender        string  `db:"gender"`
}

// Initiate database function with config from config file.
func InitDb() (dbmap *gorp.DbMap, config Config, err error) {

	config.Db.Hostname = V.GetString("database.hostname")
	config.Db.Port = V.GetInt64("database.port")
	config.Db.Username = V.GetString("database.username")
	config.Db.Password = V.GetString("database.password")
	config.Db.Database = V.GetString("database.database")

	L.Debug("Config database", "config", config)

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

// Function to insert user into database.
func InsertUser(user UserIn) (err error) {
	Dbmap, Cfg, err = InitDb()
	ErrHandler(err)
	t, err := time.Parse("02-01-2006", user.DateOfBirth)
	var q string
	ErrHandler(err)

	u := User{
		MobileNumber: user.Mobile,
		Email:        user.Email,
		Firstname:    user.Firstname,
		Lastname:     user.Lastname,
		DateOfBirth:  t,
		Gender:       user.Gender,
	}

	q = fmt.Sprintf("select count(*) from user where mobile_number = '%s'", user.Mobile)
	mobileNum, err := Dbmap.SelectInt(q)
	ErrHandler(err)
	L.Debug("Number of duplicate mobile found.", "mobileNum", mobileNum)

	q = fmt.Sprintf("select count(*) from user where email = '%s'", user.Email)
	emailNum, err := Dbmap.SelectInt(q)
	ErrHandler(err)
	L.Debug("Email of duplicate email found.", "emailNum", emailNum)

	if (mobileNum == 0) && (emailNum == 0) {
		L.Debug("1")
		err = Dbmap.Insert(&u)
		L.Debug("U variable after insert", "u", u)
		ErrHandler(err)
	} else {
		L.Debug("2")
		// duplicate found
		if mobileNum > 0 {
			err = errors.New("Duplicate mobile_number found.")
		}

		if emailNum > 0 {
			err = errors.New("Duplicate email found.")
		}

		if (emailNum > 0) && (mobileNum > 0) {
			err = errors.New("Mobile number and email must be unique.")
		}
	}

	return
}

func GetUser(user UserIn) (result UserIn, err error) {
	Dbmap, Cfg, err = InitDb()
	ErrHandler(err)

	var res UserResult
	var q string
	L.Debug("Query of getUser", "q", q)

	q = fmt.Sprintf("select * from user where mobile_number = '%s' and email = '%s'", user.Mobile, user.Email)
	err = Dbmap.SelectOne(&res, q)
	ErrHandler(err)

	L.Debug("query input", "user", user)
	L.Debug("result of database", "res", res)
	L.Debug("dob", "res", string(res.Date_of_birth))

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
		L.Debug("Content of res.Date_of_birth", "string(res.Date_of_birth)", string(res.Date_of_birth))
		L.Debug("Content of res.Date_of_birth", "res.Date_of_birth", res.Date_of_birth)
		L.Debug("Content of res.Date_of_birth", "len(res.Date_of_birth)", len(res.Date_of_birth))
		ErrHandler(err)
		result.DateOfBirth = t.Format("02-01-2006")
	}
	result.Gender = res.Gender

	return
}
