package main

import (
	"net/http"

	"gopkg.in/mgo.v2"
)

type heroesInfo struct {
	DotaName  string    `json:"name" bson:"dota_name" ` // dota name
	DotaID    int       `json:"id" bson:"id"`           // dota id
	Name      string    `json:"localized_name" bson:"name"`
	Languages Languages `json:"languages" bson:"languages"`
}

type heroJSON struct {
	Heroes []heroesInfo `json:"heroes"`
}

func handleHero(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("hero")
	query := c.Find(nil)
	var result []heroesInfo
	err := query.All(&result)
	heroes := &heroJSON{
		Heroes: result,
	}

	if err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
	}
	respond(w, r, http.StatusOK, heroes)
}
