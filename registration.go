package main

import (
	"fmt"
	"net/http"

	"errors"
	// "strconv"
	"strings"
	// "time"

	"github.com/gin-gonic/gin"
	validator "gopkg.in/asaskevich/govalidator.v10"
	log "gopkg.in/inconshreveable/log15.v2"
)

type UserIn struct {
	Mobile      string `json:"mobile" binding:"required" valid:"idnmobile,numeric"`
	Email       string `json:"email" binding:"required" valid:"email"`
	Firstname   string `json:"firstname" binding:"required" valid:"alpha"`
	Lastname    string `json:"lastname" binding:"required" valid:"alpha"`
	DateOfBirth string `json:"date_of_birth" valid:"date"`
	Gender      string `json:"gender" valid:"in(male|female|),optional"`
}

var NOT_ACCEPTABLE = gin.H{"code": "NOT_ACCEPTABLE", "message": "You are trying to request something not acceptible here."}
var NOT_FOUND = gin.H{"code": "NOT_FOUND", "message": "You are find something we can't found it here."}
var segments []string

/*
POST   /user
GET    /user/(:uid)
GET    /user/all/(:offset)/(:limit)
-----
PUT    /user/(:uid)
DELETE /user/(:uid)
*/

func RegistrationHandler(c *gin.Context) {
	segments = strings.Split(c.Param("path1"), "/")
	log.Debug(fmt.Sprintf("segments: %+v", segments))
	// log.Printf("\n%+v\n", c.Request.Method)
	// log.Printf("\n%+v\n", c.Param("path1"))
	// log.Printf("\n%+v\n", segments)
	// log.Printf("\n%+v\n", len(segments))

	switch c.Request.Method {
	case "POST":
		if strings.Compare(segments[1], "") == 0 {
			// dummyResponse(c)
			PostUserHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, NOT_ACCEPTABLE)
		}
	case "GET":
		fallthrough
	case "PUT":
		fallthrough
	case "DELETE":
		fallthrough
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, NOT_ACCEPTABLE)
	}
	// c.String(http.StatusOK, "Hello world!")
}

func PostUserHandler(c *gin.Context) {
	var input UserIn
	var err error

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	usr, err := GetUser(input)
	ErrHandler(err)
	log.Debug("Error GetUser with duplicate.")
	log.Debug("It's from query", "usr", usr)
	log.Debug("It's from input", "input", input)
	if (strings.Compare(usr.Mobile, input.Mobile) == 0) ||
		(strings.Compare(usr.Email, input.Email) == 0) {
		c.JSON(http.StatusConflict, gin.H{"code": DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", errors.New("Mobile number and or Email should be unique."))})
		return
	}

	valid, err := validator.ValidateStruct(input)
	ErrHandler(err)
	if valid {
		err = InsertUser(input)

		if err != nil {
			log.Debug("Problem in insert user", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"code": DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	} else {
		log.Debug("Problem in insert user", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func GetUserHandler(c *gin.Context) {
	var input UserIn
	var err error

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	usr, err := GetUser(input)
	ErrHandler(err)
	if (strings.Compare(usr.Mobile, input.Mobile) == 0) ||
		(strings.Compare(usr.Email, input.Email) == 0) {
		c.JSON(http.StatusConflict, gin.H{"code": DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", errors.New("Mobile number and or Email should be unique."))})
		return
	}

	err = InsertUser(input)

	if err != nil {
		log.Debug("Problem in insert user", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, input)
}
