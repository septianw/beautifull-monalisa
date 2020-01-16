package main

import (
	// "errors"
	"net/http"

	// "strings"

	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	// "github.com/spf13/viper"

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

// var L log15.Logger
// var Config *viper.Viper

var ListenAddr = "0.0.0.0:5987"

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

	// Initiate a logger
	// L = log15.New("module", "main")

	// Initiate a config
	// Config = viper.New()

}

func main() {
	ErrHandler(err)
	r := gin.Default()
	r = Router(r)

	// r := SetupRouter()
	srv := &http.Server{
		Addr:    ListenAddr,
		Handler: r,
	}
	fmt.Printf("Listening at: %s", ListenAddr)

	// srv.ListenAndServe()

	// gracefull shutdown procedure

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
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
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		// os.Remove("/tmp/shinyRuntimeFile")
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

func Router(r *gin.Engine) *gin.Engine {
	// r.Use(cors.Default())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
		AllowHeaders:     []string{"Authorization"},
		AllowCredentials: true,
	}))

	// r.Use(func(c *gin.Context) {
	// 	var h Head
	// 	log.Println("1")

	// 	log.Println(c.GetHeader("Authorization"))

	// 	if err = c.ShouldBindHeader(&h); err != nil {
	// 		ErrHandler(err)
	// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": INPUT_VALIDATION_FAIL,
	// 			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", errors.New("x-terminal header not found. access forbidden.").Error())})
	// 		c.Abort()
	// 		return
	// 	}

	// 	// check authorization
	// 	if strings.Compare(h.Authorization, "") == 0 {
	// 		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": FORBIDDEN_CODE,
	// 			"message": fmt.Sprintf("FORBIDDEN: %s", errors.New("Authorization required. Please login first.").Error())})
	// 		return
	// 	} else if strings.Compare(strings.Split(h.Authorization, " ")[1], apikey) == 0 {
	// 		c.Next()
	// 		return
	// 	}

	// 	return
	// })

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello world!")
	})

	r.Any("/api/v1/user/*path1", RegistrationHandler)

	r.GET("/ws", func(c *gin.Context) {
		WsHandler(c.Writer, c.Request)
	})

	return r
}

// Set websocket cors whitelist here.
func checkWhitelist(r *http.Request) bool {
	return true
}

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
