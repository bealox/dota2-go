package main

import (
	"net/http"

	"gopkg.in/mgo.v2"
)

type abilityInfo struct {
	Name   string `json:"name" bson:"name"`
	DotaID string `json:"id" bson:"id"`
}

type abilityJSON struct {
	Abilities []abilityInfo `json:"abilities"`
}

func handleAbility(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("ability")
	var result []abilityInfo
	query := c.Find(nil)
	err := query.All(&result)
	ability := &abilityJSON{
		Abilities: result,
	}

	if err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
	}

	respond(w, r, http.StatusOK, ability)
}
