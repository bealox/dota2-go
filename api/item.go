package main

import (
	"net/http"

	"gopkg.in/mgo.v2"
)

type (
	itemInfo struct {
		DotaID     int       `json:"id" bson:"id"`                   // the ID used to identify the item in the api
		DotaName   string    `json:"name" bson:"dota_name"`          // the code name of the item
		Cost       int       `json:"cost" bson:"cost"`               // the gold cost of the item
		SecretShop int       `json:"secret_shop" bson:"secret_shop"` // 1 if is available, 0 otherwise
		SideShop   int       `json:"side_shop" bson:"side_shop"`     // 1 if is available, 0 otherwise
		Recipt     int       `json:"recipe" bson:"recipe"`           // 1 if is available, 0 otherwise
		Name       string    `json:"localized_name" bson:"name"`     // if a lang specified, this will show the in game name of the item for that language
		Languages  Languages `json:"languages" bson:"languages"`
	}

	itemJSON struct {
		Items []itemInfo `json:"items"`
	}
)

func handleItem(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("item")
	query := c.Find(nil)
	var result []itemInfo
	err := query.All(&result)
	if err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
	}
	items := &itemJSON{
		Items: result,
	}
	respond(w, r, http.StatusOK, items)
}
