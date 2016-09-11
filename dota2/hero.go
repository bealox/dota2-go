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
	//Hero format
	Hero struct {
		JSON *HeroContent
		DB   *mgo.Database
	}

	//HeroContent is the json format for hero.
	HeroContent struct {
		Result struct {
			Heroes []*gHeroesInfo `json:"heroes" bson:"heroes"`
			Status int            `json:"status" bson:"status"`
			Count  int            `json:"count" bson:"count"`
		} `json:"result" bson:"result"`
	}

	gHeroesInfo struct {
		DotaName string `json:"name" bson:"dota_name" ` // dota name
		DotaID   int    `json:"id" bson:"id"`           // dota id
		Name     string `json:"localized_name" bson:"name"`
	}
)

//GetHero allow user to call it
func GetHero(db *mgo.Database) *Hero {
	return &Hero{
		DB: db,
	}
}

func (h *Hero) getURLs() map[string]string {
	urls := map[string]string{}
	baseURL := "http://api.steampowered.com/IEconDOTA2_570/GetHeroes/v1"
	for _, lang := range steamDetails.languages {
		para := url.Values{}
		para.Add("key", steamDetails.key)
		para.Add("language", lang)
		url := baseURL + "?" + para.Encode()
		urls[lang] = url
	}
	log.Println("ur ", urls)
	return urls
}

func (h *Hero) getJSON() error {
	urls := h.getURLs()

	for lang, url := range urls {
		body, err := http.Get(url)
		if err != nil {
			return err
		}
		err = json.NewDecoder(body.Body).Decode(&h.JSON)
		if err != nil {
			log.Fatalln("unable to decode JSON ", err)
		}
		// defer body.Body.Close()
		h.parseJSON(lang)
	}
	return errors.New("unable get urls")
}

func (h *Hero) parseJSON(lang string) {
	collection := h.DB.C("hero")
	count, err := collection.Count()

	if err != nil {
		log.Fatalln("Unable to get collection count ", err)
	}

	if count == 0 {
		for _, i := range h.JSON.Result.Heroes {
			err = collection.Insert(i)
			if err != nil {
				log.Fatalln("Unable insert into Mongo ", err)
			}
			//TODO: should I run stroeHeroImage in goroutine?
			err = storeHeroImage(i.DotaName, "./Hero/")
			if err != nil {
				log.Fatalln("Unable store image ", err)
			}
		}
	} else {
		for _, i := range h.JSON.Result.Heroes {
			count, err = collection.Find(bson.M{"id": i.DotaID}).Count()
			if err != nil {
				log.Fatalln("unable to find existing ", err)
			}
			if count == 0 {
				collection.Insert(i)
				err = storeHeroImage(i.DotaName, "./Hero/")
				if err != nil {
					log.Fatalln("Unable store image ", err)
				}
			}
		}
	}

	/** Setting up languages **/
	for _, i := range h.JSON.Result.Heroes {
		err = collection.Update(bson.M{"id": i.DotaID}, bson.M{"$set": bson.M{"languages." + lang: i.Name}})
		if err != nil {
			log.Fatalln("Unable insert into Mongo ", err)
		}
	}

}

func storeHeroImage(dotaName string, dir string) error {
	//small horizontal portrait - 59x33px sb.png
	//large horizontal portrait - 205x11px lg.png
	//full quality horizontal portrait - 256x114px full.png
	//full quality vertical portrait - 234x272px vert.jpg

	extentions := []string{"sb.png", "lg.png", "full.png", "vert.jpg"}
	apiURL := "http://cdn.dota2.com/apps/dota2/images/heroes/"
	fileName := strings.Replace(dotaName, "npc_dota_hero_", "", -1)

	for _, ext := range extentions {
		path := dir + dotaName + "_" + ext
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

func (h *Hero) run() {
	h.getJSON()
}
