package main

import (
	"encoding/json"
	"errors"
	"strings"

	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	log "gopkg.in/inconshreveable/log15.v2"
)

type header map[string]string
type headers []header
type payload struct {
	Method string
	Url    string
	Body   io.Reader
}
type expectation struct {
	Code int
	Body string
}

type quest struct {
	pload  payload
	heads  headers
	expect expectation
}
type quests []quest

var upgrader = websocket.Upgrader{}

// test http tool start
func getArm() (*gin.Engine, *httptest.ResponseRecorder) {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	Router(router)

	recorder := httptest.NewRecorder()
	return router, recorder
}

func handleErr(err error, t *testing.T) {
	if err != nil {
		if t != nil {
			t.Errorf(fmt.Sprintf("%+v", err))
			t.Fail()
		} else {
			log.Crit(err.Error())
		}
	}
}

func doTheTest(load payload, heads headers) *httptest.ResponseRecorder {
	var router, recorder = getArm()

	req, err := http.NewRequest(load.Method, load.Url, load.Body)
	log.Debug(fmt.Sprintf("%+v", req))
	// log.Printf("%+v", req)
	handleErr(err, nil)

	if len(heads) != 0 {
		for _, head := range heads {
			for key, value := range head {
				req.Header.Set(key, value)
			}
		}
	}
	router.ServeHTTP(recorder, req)

	return recorder
}

func SetupRouter() *gin.Engine {
	return gin.New()
}

// test http tool end

// make dummy websocket server here
func GenServer(handler http.Handler) (ws *websocket.Conn, err error) {
	s := httptest.NewServer(handler)
	defer s.Close()

	// convert http://localhost to ws://localhost
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// connect to the server
	ws, _, err = websocket.DefaultDialer.Dial(u, nil)
	// t.Logf("ws: %+v, res: %+v, err: %+v\n", ws, res, err)

	return
}

func TestErrHandling(t *testing.T) {
	var err error

	err = errors.New("Error testing")
	os.Setenv("RUNMODE", "development")
	ErrHandler(err)

	err = errors.New("Error testing")
	os.Setenv("RUNMODE", "testing")
	ErrHandler(err)

	err = errors.New("Error testing")
	os.Setenv("RUNMODE", "production")
	ErrHandler(err)

}

func TestFirstInit(t *testing.T) {
	os.Setenv("RUNMODE", "testing")
	dbmap, config, err := InitDb()
	t.Log(err)
	t.Log(dbmap)
	t.Log(config)
	if err != nil {
		t.Fail()
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Db.Username,
		config.Db.Password,
		config.Db.Hostname,
		config.Db.Port,
		config.Db.Database,
	))
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	rows, err := db.Query("select * from user")
	handleErr(err, t)

	cols, _ := rows.Columns()
	t.Logf("%+v", cols)
}

func TestGetCredential(t *testing.T) {
	ws, err := GenServer(http.HandlerFunc(WsHandler))
	handleErr(err, t)
	defer ws.Close()

	mType, msg, err := ws.ReadMessage()
	t.Logf("type: %+v, msg: %+v, err: %+v]n", mType, string(msg), err)
	handleErr(err, t)

	if strings.Compare(string(msg), apikey) != 0 {
		t.Fail()
	}
}

func PostUser(userin UserIn, expect expectation, t *testing.T) bool {
	var err error
	Dbmap, _, err = InitDb()
	handleErr(err, t)
	err = Dbmap.TruncateTables()
	handleErr(err, t)

	os.Setenv("RUNMODE", "development")

	useraddJson, err := json.Marshal(userin)
	handleErr(err, t)
	t.Log(string(useraddJson))

	NewUser := strings.NewReader(string(useraddJson))

	q := quest{
		payload{"POST", "/api/v1/user/", NewUser},
		headers{
			header{"Authorization": "Bearer c66f2005-b556-47e9-8086-328a354e6064"},
		},
		expect,
	}

	rec := doTheTest(q.pload, q.heads)
	t.Log(rec)
	t.Log(rec.Body.String())

	return assert.Equal(t, expect.Code, rec.Code) &&
		assert.Equal(t, expect.Body, strings.TrimSpace(rec.Body.String()))
}

func TestPostUserPositive(t *testing.T) {
	var userin UserIn

	userin.Mobile = "08123456789"
	userin.Email = "user@network.net"
	userin.Firstname = "Maya"
	userin.Lastname = "Lauren"
	userin.DateOfBirth = "20-03-1986"
	userin.Gender = "female"

	result, err := GetUser(userin)
	handleErr(err, t)
	t.Logf("%+v", result)
	t.Logf("%+v", result.Mobile)

	jsonUser, err := json.Marshal(result)
	handleErr(err, t)

	expect := expectation{201, string(jsonUser)}

	assert.True(t, PostUser(userin, expect, t))
}

// func TestPostUserNegativeAll(t *testing.T) {
// 	var userin UserIn

// 	userin.Mobile = "098712345678"
// 	userin.Email = "usernetworknet"
// 	userin.Firstname = ""
// 	userin.Lastname = ""
// 	userin.DateOfBirth = "20-03-1024"
// 	userin.Gender = "apache helicopter"

// 	result, err := GetUser(userin)
// 	handleErr(err, t)
// 	t.Logf("%+v", result)
// 	t.Logf("%+v", result.Mobile)

// 	jsonUser, err := json.Marshal(result)
// 	handleErr(err, t)

// 	expect := expectation{400, string(jsonUser)}

// 	assert.True(t, PostUser(userin, expect, t))
// }

func TestGetUser(t *testing.T) {
	var userin UserIn
	Dbmap, _, err = InitDb()
	handleErr(err, t)
	err = Dbmap.TruncateTables()
	handleErr(err, t)

	os.Setenv("RUNMODE", "development")

	err = InsertUser(UserIn{
		Mobile:      "08123456789",
		Email:       "user@network.net",
		Firstname:   "Maya",
		Lastname:    "Lauren",
		DateOfBirth: "20-03-1986",
		Gender:      "female",
	})
	handleErr(err, t)

	userin.Mobile = "08123456789"
	userin.Email = "user@network.net"
	userin.Firstname = "Maya"
	userin.Lastname = "Lauren"
	userin.DateOfBirth = "20-03-1986"

	userout, err := GetUser(userin)
	handleErr(err, t)

	log.Debug("Result of getUser", "userout", userout)
}
