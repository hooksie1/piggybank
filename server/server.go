package server

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var dbName = os.Getenv("DATABASE_PATH")

var masterPass []byte

type errHandler func(http.ResponseWriter, *http.Request) error

func checkEnv() bool {
	if os.Getenv("DATABASE_PATH") == "" {
		return false
	}

	return true
}

func Serve() {
	if !checkEnv() {
		log.Println("You must set the database path with env \"DATABASE_PATH\"")
		os.Exit(1)
	}

	apiPath := "/api"
	initPath := "/init"

	router := mux.NewRouter().StrictSlash(true)
	apiRouter := router.PathPrefix(apiPath).Subrouter().StrictSlash(true)
	initRouter := router.PathPrefix(initPath).Subrouter().StrictSlash(true)

	//apiRouter.HandleFunc("/", index).Methods("GET")
	initRouter.Handle("/unlock", errHandler(unlockSystem)).Methods("POST")
	initRouter.Handle("/initialize", errHandler(initialize)).Methods("POST")
	apiRouter.Handle("/password", errHandler(addPass)).Methods("POST")
	apiRouter.Handle("/password", errHandler(getPass)).Methods("GET")
	apiRouter.Handle("/password", errHandler(deletePass)).Methods("DELETE")
	apiRouter.Handle("/status", errHandler(getStatus)).Methods("GET")
	apiRouter.Handle("/user/{userName}", errHandler(createUser)).Methods("POST")
	apiRouter.Handle("/user/{userName}", errHandler(deleteUser)).Methods("DELETE")
	apiRouter.Handle("/backup", errHandler(Backup)).Methods("POST")

	initRouter.Use(logger)

	apiRouter.Use(logger)
	apiRouter.Use(authentication)
	apiRouter.Use(checkDB)

	log.Fatal(http.ListenAndServe(":8080", router))
}
