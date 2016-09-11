package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (

	//Ability format
	Ability struct {
		JSON gAbilityInfo
		DB   *mgo.Database
	}

	gAbility struct {
		Name   string `json:"name" bson:"name"`
		DotaID string `json:"id" bson:"id"`
	}

	gAbilityInfo struct {
		Abilities []gAbility `json:"abilities"`
	}
)

//GetAbility allow user to call it
func GetAbility(db *mgo.Database) *Ability {
	return &Ability{
		DB: db,
	}
}

func (a *Ability) getURLs() map[string]string {
	urls := map[string]string{}
	urls["en-us"] = "https://raw.githubusercontent.com/kronusme/dota2-api/master/data/abilities.json"
	return urls
}

func (a *Ability) getJSON() error {
	urls := a.getURLs()

	for _, url := range urls {
		body, err := http.Get(url)
		if err != nil {
			log.Fatalln("unable to get ability JSON err:", err)
		}
		var result gAbilityInfo
		err = json.NewDecoder(body.Body).Decode(&result)
		if err != nil {
			log.Fatalln("unable to decode ability JSON err:", err)
		}
		a.JSON = result
		a.parseJSON()
	}
	return errors.New("unable get JSON for ")
}

func (a *Ability) parseJSON() {
	collection := a.DB.C("ability")
	for _, i := range a.JSON.Abilities {
		count, err := collection.Find(bson.M{"id": i.DotaID}).Count()
		if err != nil {
			log.Fatalln("unable to match ability id in DB err:", err)
		}
		if count == 0 {
			collection.Insert(i)
		}
	}
}

func (a *Ability) run() {
	a.getJSON()
}
