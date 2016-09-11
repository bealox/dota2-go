package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gopkg.in/mgo.v2"

	"github.com/gorilla/pat"
)

var port string

func init() {
	flag.StringVar(&port, "port", "8080", "port for the dota2 API")
}

func main() {

	flag.Parse()

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprint(w, "Hello World")
	// })
	db, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalln("unable to connect to mongodb, err:", err)
	}
	defer db.Close()

	r := pat.New()
	r.Get("/heroes", withCORS(withVars(withData(withAPIKey(handleHero), db))))
	r.Get("/items", withCORS(withVars(withData(withAPIKey(handleItem), db))))
	r.Get("/abililities", withCORS(withVars(withData(withAPIKey(handleAbility), db))))

	r.Get("/", handleHome)
	http.Handle("/", r)

	log.Println("launching api :" + port)
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello to Dota2")
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fn(w, r)
	}
}

func withVars(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		OpenVars(r)
		defer CloseVars(r)
		fn(w, r)
	}
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidAPI(r.URL.Query().Get("key")) {
			respondErr(w, r, http.StatusUnauthorized, "Invalid API Key")
			return
		}
		fn(w, r)
	}
}

func isValidAPI(key string) bool {
	if key == "qwer" {
		return true
	}
	return false
}

func withData(fn http.HandlerFunc, db *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		thisDb := db.Copy()
		defer thisDb.Close()
		SetVar(r, "db", db.DB("dota2"))
		fn(w, r)
	}
}
