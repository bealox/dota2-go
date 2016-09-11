package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	//Item JSON format
	Item struct {
		JSON *ItemContent
		DB   *mgo.Database
	}

	//ItemContent based on the JSON sturcture
	ItemContent struct {
		Result struct {
			Items  []gItemInfo `json:"items"`
			Status int         `json:"status"`
		} `json:"result"`
	}

	gItemInfo struct {
		DotaID     int    `json:"id" bson:"id"`                   // the ID used to identify the item in the api
		DotaName   string `json:"name" bson:"dota_name"`          // the code name of the item
		Cost       int    `json:"cost" bson:"cost"`               // the gold cost of the item
		SecretShop int    `json:"secret_shop" bson:"secret_shop"` // 1 if is available, 0 otherwise
		SideShop   int    `json:"side_shop" bson:"side_shop"`     // 1 if is available, 0 otherwise
		Recipt     int    `json:"recipe" bson:"recipe"`           // 1 if is available, 0 otherwise
		Name       string `json:"localized_name" bson:"name"`     // if a lang specified, this will show the in game name of the item for that language
	}
)

//GetItem allow user to call it
func GetItem(db *mgo.Database) *Item {
	return &Item{
		DB: db,
	}
}

func (i *Item) getURLs() map[string]string {
	urls := map[string]string{}
	baseURL := "https://api.steampowered.com/IEconDOTA2_570/GetGameItems/V001/?"
	for _, lang := range steamDetails.languages {
		para := url.Values{}
		para.Add("key", steamDetails.key)
		para.Add("language", lang)
		url := baseURL + "?" + para.Encode()
		urls[lang] = url
	}
	log.Println("urls: ", urls)
	return urls
}

func (i *Item) getJSON() error {
	urls := i.getURLs()

	for lang, url := range urls {
		body, err := http.Get(url)
		if err != nil {
			return err
		}
		err = json.NewDecoder(body.Body).Decode(&i.JSON)
		defer body.Body.Close()
		if err != nil {
			log.Fatalln("unable to decode JSON ", err)
		}
		i.parseJSON(lang)
	}
	return errors.New("unable get urls")
}

func (i *Item) parseJSON(lang string) {
	collection := i.DB.C("item")
	count, err := collection.Count()

	if err != nil {
		log.Fatalln("Unable to get collection count ", err)
	}

	if count == 0 {
		for _, i := range i.JSON.Result.Items {
			err = collection.Insert(i)
			if err != nil {
				log.Fatalln("Unable insert into Mongo ", err)
			}
			//TODO: should I run stroeHeroImage in goroutine?
			err = storeItemImage(i.DotaName, "./Item/")
			if err != nil {
				log.Fatalln("Unable store image ", err)
			}

			err = collection.Update(bson.M{"id": i.DotaID}, bson.M{"$set": bson.M{"languages." + lang: i.Name}})
			if err != nil {
				log.Fatalln("Unable insert into Mongo ", err)
			}
		}
	} else {
		for _, i := range i.JSON.Result.Items {
			count, err = collection.Find(bson.M{"id": i.DotaID}).Count()
			if err != nil {
				log.Fatalln("unable to find existing ", err)
			}
			if count == 0 {
				collection.Insert(i)
				err = storeItemImage(i.DotaName, "./Item/")
				if err != nil {
					log.Fatalln("Unable store image ", err)
				}
			}

			err = collection.Update(bson.M{"id": i.DotaID}, bson.M{"$set": bson.M{"languages." + lang: i.Name}})
			if err != nil {
				log.Fatalln("Unable insert into Mongo ", err)
			}
		}
	}
}

func storeItemImage(dotaName string, dir string) error {
	//large horizontal portrait - 205x11px lg.png

	extentions := []string{"lg.png"}
	apiURL := "http://cdn.dota2.com/apps/dota2/images/items/"
	fileName := strings.Replace(dotaName, "item_", "", -1)

	for _, ext := range extentions {
		path := dir + fileName + "_" + ext
		if _, err := os.Stat(path); os.IsNotExist(err) {
			resp, err := http.Get(apiURL + fileName + "_" + ext)

			if err != nil {
				return err
			}

			file, err := os.Create(path)
			if err != nil {
				return err
			}

			_, err = io.Copy(file, resp.Body)
			resp.Body.Close()
			file.Close()
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (i *Item) run() {
	i.getJSON()
}
