package main

import (
	"errors"
	"net/http"

	// "strings"

	"context"
	"fmt"

	// "log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	validator "gopkg.in/asaskevich/govalidator.v10"
	"gopkg.in/gorp.v1"
)

type Head struct {
	Authorization string `header:"Authorization"`
}

// change this acording to documentation
const INPUT_VALIDATION_FAIL = 2110
const DATABASE_EXEC_FAIL = 2200
const FORBIDDEN_CODE = 2103

// put your dummy apikey here
var apikey = "c66f2005-b556-47e9-8086-328a354e6064"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

var Dbmap *gorp.DbMap
var Cfg Config
var err error

var L log15.Logger

var V *viper.Viper

// Default listening address
var ListenAddr = "0.0.0.0:5987"

/*
	This is initiate function that always run at the first time before anything else.
*/
func init() {
	// put custom validator to assert mobile input by user
	validator.CustomTypeTagMap.Set("idnmobile", func(i interface{}, context interface{}) bool {
		var valid = false
		var d string
		var validPrefix = []string{"811", "812", "813", "814", "815", "816", "817", "818", "819", "838", "852", "853", "855", "856", "858", "859", "878", "896", "897", "898", "899"}
		var prefixStart = 0
		var prefixEnd = 3

		switch v := i.(type) {
		case string:
			d = v
		default:
			return valid
		}

		log15.Debug("input value", "mobile", d)

		if d[0:1] == "0" {
			prefixStart = 1
			prefixEnd = 4
		}

		for _, b := range validPrefix {
			if (b == d[prefixStart:prefixEnd]) && (len(d) > 9) && (len(d) < 13) {
				valid = true
			}
		}

		return valid
	})

	// put custom validator to assert date input by user.
	validator.CustomTypeTagMap.Set("date", func(i interface{}, context interface{}) bool {
		var valid = false
		var d string
		var err error
		var validDate, validMonth, validYear bool
		var now = time.Now()

		switch v := i.(type) {
		case string:
			d = v
		default:
			return valid
		}

		// FIXME: nasty workaround to skip empty date. Since i can't enable skipping.
		if strings.Compare(d, "") == 0 {
			return true
		}

		ds := strings.Split(d, "-")
		log15.Debug("Splitted date from user", "ds", ds)
		date, err := strconv.Atoi(ds[0])
		ErrHandler(err)
		if err != nil {
			return valid
		}
		month, err := strconv.Atoi(ds[1])
		ErrHandler(err)
		if err != nil {
			return valid
		}
		year, err := strconv.Atoi(ds[2])
		ErrHandler(err)
		if err != nil {
			return valid
		}

		if (date < 31) && (date > 1) {
			validDate = true
		}

		if (month < 13) && (month > 0) {
			validMonth = true
		}

		if (year < now.Year()) && (year > now.Year()-150) {
			validYear = true
		}

		if validDate && validMonth && validYear {
			valid = true
		}

		return valid
	})

	// Initiate a config
	V = viper.New()
	V.SetConfigName("config")
	V.SetConfigType("toml")
	V.AddConfigPath(".")
	if strings.Compare(os.Getenv("RUNMODE"), "testing") == 0 {
		V.AddConfigPath("$HOME")
	}

	err = V.ReadInConfig()
	ErrHandler(err)
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// log.Fataln("No config file found.")
			L.Crit("No config file found.")
			os.Exit(2)
			// Config file not found; ignore error if desired
		} else {
			// log.Fatalln(err)
			L.Crit("Error at read config", "err.Error()", err.Error())
			os.Exit(2)
			// Config file was found but another error was produced
		}
	}

	ListenAddr = fmt.Sprintf("%s:%d", V.GetString("server.bind"), V.GetInt64("server.port"))

	// Initiate a logger
	L = log15.New("module", "main")
	L.SetHandler(log15.MultiHandler(
		log15.StreamHandler(os.Stderr, log15.LogfmtFormat()),
		log15.FilterHandler(func(r *log15.Record) bool {
			return r.Lvl == log15.LvlError
		}, log15.Must.FileHandler("error.json", log15.JsonFormat())),
		log15.FilterHandler(func(r *log15.Record) bool {
			var out bool
			switch os.Getenv("RUNMODE") {
			case "development":
				out = (r.Lvl == log15.LvlCrit) || (r.Lvl == log15.LvlError) ||
					(r.Lvl == log15.LvlWarn) || (r.Lvl == log15.LvlInfo) || (r.Lvl == log15.LvlDebug)
			case "testing":
				out = (r.Lvl == log15.LvlCrit) || (r.Lvl == log15.LvlError) ||
					(r.Lvl == log15.LvlWarn) || (r.Lvl == log15.LvlInfo)
			case "production":
				out = (r.Lvl == log15.LvlCrit) || (r.Lvl == log15.LvlError) ||
					(r.Lvl == log15.LvlWarn)
			default:
				out = (r.Lvl == log15.LvlCrit)
			}
			return out
		}, log15.StreamHandler(os.Stdout, log15.LogfmtFormat())),
	))
}

func main() {
	ErrHandler(err)
	r := gin.Default()
	r = Router(r)

	srv := &http.Server{
		Addr:    ListenAddr,
		Handler: r,
	}
	fmt.Printf("\nListening at: %s\n", ListenAddr)

	// gracefull shutdown procedure
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			L.Crit("Fail to listen", "err", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	L.Warn("Server Shutting down.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		L.Crit("Server fail to start", "err", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		// do everything we need to shutdown here
		L.Info("5 seconds timeout.")
	}
	L.Info("Server exiting")
}

// Initiate route function
func Router(r *gin.Engine) *gin.Engine {
	// Enable cors middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
		AllowHeaders:     []string{"Authorization"},
		AllowCredentials: true,
	}))

	r.Use(func(c *gin.Context) {
		var h Head

		L.Debug("Full path at middleware", "c.FullPath()", c.FullPath())

		// if (strings.Compare(c.FullPath(), "/ws") == 0) || (strings.Compare(c.FullPath()[:3], "/ui") == 0) {
		// 	c.Next()
		// 	return
		// }
		L.Debug("Get header at middleware", `c.GetHeader("Authorization")`, c.GetHeader("Authorization"))
		// L.Debug("Full path of api", "c.FullPath()[:4]", c.FullPath()[:4])
		// L.Debug("result of compare", "s.compare", strings.Compare(c.FullPath()[:4], "/api"))

		// bind header to header variable and fail if can't bind
		if strings.Compare(c.FullPath()[1:3], "/api") == 0 {
			if err = c.ShouldBindHeader(&h); err != nil {
				ErrHandler(err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": INPUT_VALIDATION_FAIL,
					"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", errors.New("Authorization header not found. access forbidden.").Error())})
				c.Abort()
				return
			}

			// check authorization
			if strings.Compare(h.Authorization, "") == 0 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": FORBIDDEN_CODE,
					"message": fmt.Sprintf("FORBIDDEN: %s", errors.New("Authorization required. Please login first.").Error())})
				return
			} else if strings.Compare(strings.Split(h.Authorization, " ")[1], apikey) == 0 {
				c.Next()
				return
			}
		}

		c.Next()
		return
	})

	// this is placeholder for UI
	// r.StaticFS("/ui")
	r.StaticFS("/ui", FS(false))
	// r.GET("/", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "Hello world!")
	// })

	// this is the only endpoint for backend.
	r.Any("/api/v1/user/*path1", RegistrationHandler)

	// This websocket interface used for transmit apikey
	// from server to javascript client
	r.GET("/ws", func(c *gin.Context) {
		WsHandler(c.Writer, c.Request)
	})

	return r
}

// Set websocket cors whitelist here.
func checkWhitelist(r *http.Request) bool {
	return true
}

// Websocket handler
func WsHandler(w http.ResponseWriter, r *http.Request) {
	wsUpgrader.CheckOrigin = checkWhitelist

	c, err := wsUpgrader.Upgrade(w, r, nil)
	ErrHandler(err)

	// this websocket purpose is only to send apikey to client
	// after apikey sent, close the socket.
	err = c.WriteMessage(1, []byte(apikey))
	ErrHandler(err)
	c.Close()
}
