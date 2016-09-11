package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/**
INFO:
STEAMID32 is the account ID in this case playerID
STEAMID64 is the steam ID , you need it to get player info
STEAMID64 - 76561197960265728 = STEAMID32
STEAMID32 + 76561197960265728 = STEAMID64

EE(43276219): http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=<key>&steamids=76561198003541947
Mushi(89871557): http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=<key>&steamids=76561198050137285
*/

type (
	// player
	player struct {
		DB   *mgo.Database
		JSON *PlayerJSON
	}

	gPlayerContent struct {
		AccountID                string `json:"omitempty" bson:"_id"`
		SteamID                  string `json:"steamid" bson:"steamid"`
		CommunityVisibilityState uint8  `json:"communityvisibilitystate" bson:"communityvisibilitystate"`
		/**
			communityvisibilitystate
			1 Private
			2 Friend Only
			3 Friends of Friends
			4 Users Only
			5 Public
		**/
		ProfileState     uint8  `json:"profilestate" bson:"profilestate"`
		VERIFIEDNAME     string `json:"omitempty" bson:"verifiedname"`
		NickName         string `json:"personaname" bson:"personaname"`
		LastLogOff       int    `json:"lastlogoff" bson:"lastlogoff"`
		ProfileURL       string `json:"profileurl" bson:"profileurl"`
		Avatar           string `json:"avatar" bson:"avatar"`
		AvatarMedium     string `json:"avatarmedium" bson:"avatarmedium"`
		AvatarFull       string `json:"avatarfull" bson:"avatarfull"`
		PersonState      uint8  `json:"personastate" bson:"personastate"`
		RealName         string `json:"realname" bson:"realname"`
		PrimaryClanID    string `json:"primaryclanid" bson:"primaryclanid"`
		TimeCreated      uint   `json:"timecreated" bson:"timecreated"`
		PersonStateFlags uint8  `json:"personastateflags" bson:"personastateflags"`
	}

	//PlayerJSON strct is to retrieve JSON from the server
	PlayerJSON struct {
		Response struct {
			Players []*gPlayerContent `json:"players" bson:"players"`
		} `json:"response" bson:"response"`
	}
)

//GetPlayer use this to get player struct
func GetPlayer(db *mgo.Database) *player {
	return &player{
		DB: db,
	}
}

func (p *player) processID(accountID int64) (accountid string, steamid string) {
	steamID := 76561197960265728 + accountID
	steamIDString := strconv.FormatInt(steamID, 10)
	accountIDString := strconv.FormatInt(accountID, 10)
	return accountIDString, steamIDString
}

func (p *player) getURL(steamIDString string) string {
	baseURL := "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002"
	para := url.Values{}
	para.Add("key", steamDetails.key)
	para.Add("steamids", steamIDString)
	url := baseURL + "?" + para.Encode()
	return url
}

//GetPlayer details from API
func (p *player) getPlayer(accountID int) {

	accountIDString, steamIDString := p.processID(int64(accountID))

	url := p.getURL(steamIDString)
	body, err := http.Get(url)
	if err != nil {
		log.Fatalln("unable to get data from API, err", err)
	}
	err = json.NewDecoder(body.Body).Decode(&p.JSON)
	if err != nil {
		log.Fatalln("unable to decode data from JSON, err", err)
	}
	// defer body.Body.Close()
	if p.JSON.Response.Players[0].SteamID != "" {
		p.parseJSON(accountIDString)
	}

}

func (p *player) parseJSON(accountID string) {
	collection := p.DB.C("player")
	player := p.JSON.Response.Players[0]

	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"lastlogoff": player.LastLogOff, "avatar": player.Avatar,
			"avatarmedium": player.AvatarMedium, "avatarfull": player.AvatarFull}},
		Upsert:    false,
		Remove:    false,
		ReturnNew: true,
	}

	_, err := collection.Find(bson.M{"_id": accountID}).Apply(change, nil)
	if err == mgo.ErrNotFound {
		player.AccountID = accountID
		collection.Insert(player)
	}
}
