package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//TODO: How can I update LeagueID? leagueIDs grows overtime.Maybe I can insert the whole thing and fetch it later

/**
INFO:
STEAMID32 is the account ID in this case playerID
STEAMID64 is the steam ID , you need it to get player info
STEAMID64 - 76561197960265728 = STEAMID32
STEAMID32 + 76561197960265728 = STEAMID64

Funtaic http://api.steampowered.com/IDOTA2Match_570/GetTeamInfoByTeamID/v1?key=&start_at_team_id=350190&teams_requested=1

**/

type (
	//Team format
	Team struct {
		DB   *mgo.Database
		JSON *TeamContent
		// TeamID int64 //This is needed to get a team ID

	}

	gTeamContent struct {
		TeamID      string      `json:"team_id" bson:"_id"`
		Name        string      `json:"name"`
		Tag         string      `json:"tag"`
		CreatedBy   int64       `json:"time_created"`
		Rating      interface{} `json:"rating"`
		Logo        int64       `json:"logo"`
		LogoSponsor int64       `json:"logo_sponsor"`
		CountryCode string      `json:"country_code"`
		URL         string      `json:"url"`
		Player0     int         `json:"player_0_account_id" bson:"player1_id"`
		Player1     int         `json:"player_1_account_id" bson:"player2_id"`
		Player2     int         `json:"player_2_account_id" bson:"player3_id"`
		Player3     int         `json:"player_3_account_id" bson:"player4_id"`
		Player4     int         `json:"player_4_account_id" bson:"player5_id"`
		SubPlayer   int         `json:"player_5_account_id" bson:"subplayer_id"` //subtitude
		Admin       int         `json:"admin_account_id" bson:"admin_player_id"`
	}

	//TeamContent json structure
	TeamContent struct {
		Result struct {
			Teams []gTeamContent `json:"teams"`
		} `json:"result"`
	}

	//TeamLogoJSON Getting Team logo
	TeamLogoJSON struct {
		Data struct {
			FileName string `json:"filename"`
			URL      string `json:"url"`
			Size     int    `json:"size"`
		} `json:"data"`
	}
)

//GetTeam allow user to call it
func GetTeam(db *mgo.Database) *Team {
	return &Team{
		DB: db,
	}
}

//getURL only give you show one team at a time, and team id is required
func (t *Team) getURL(teamID string) map[string]string {
	urls := map[string]string{}
	v := url.Values{}
	v.Add("key", steamDetails.key)
	v.Add("start_at_team_id", teamID)
	v.Add("teams_requested", "1")
	urls["en-us"] = "https://api.steampowered.com/IDOTA2Match_570/GetTeamInfoByTeamID/v001/?" + v.Encode()
	return urls
}

func (t *Team) getTeam(teamID int64) error {
	if teamID == 0 {
		return errors.New("TeamID cannot be zero in the Team struct")
	}

	teamIDString := strconv.FormatInt(teamID, 10)

	url := t.getURL(teamIDString)["en-us"]
	resp, err := http.Get(url)
	if err != nil {
		log.Println("unable to get JSON for team, err: ", err)
		return err
	}

	var result TeamContent
	err = json.NewDecoder(resp.Body).Decode(&result)
	defer resp.Body.Close()
	if err != nil {
		log.Println("unable to decode json, err:", err)
		return err
	}
	t.JSON = &result
	t.parseJSON(teamIDString)
	return nil
}

func getTeamLogo(idString int64, fileName string) {

	v := url.Values{}
	v.Add("ugcid", strconv.FormatInt(idString, 10))
	v.Add("key", steamDetails.key)
	v.Add("appid", "570")
	api := "http://api.steampowered.com/ISteamRemoteStorage/GetUGCFileDetails/v1/?" + v.Encode()
	log.Println(api)
	resp, err := http.Get(api)
	if err != nil {
		log.Fatalln("unable to fer json details for team logo, err:", err)
	}
	defer resp.Body.Close()

	var document *TeamLogoJSON
	if err = json.NewDecoder(resp.Body).Decode(&document); err != nil {
		log.Fatalln("unable to decode team json logo, err:", err)
	}

	rep, err := http.Get(document.Data.URL)
	if err != nil {
		log.Fatalln("unable to logo ", err)
	}
	defer rep.Body.Close()

	dir := "./Team/"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	file := dir + fileName + ".png"
	// if team logo doesn't exist then create one
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file, err := os.Create(file)
		defer file.Close()
		if err != nil {
			log.Fatalln("unable to create team logo, err:", err)
		}
		_, err = io.Copy(file, rep.Body)
		if err != nil {
			log.Fatalln("unable to copy file for team logo, err:", err)
		}
	}
}

func (t *Team) parseJSON(teamID string) {
	collection := t.DB.C("team")
	team := t.JSON.Result.Teams[0]

	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"player1_id":      team.Player0,
				"player2_id":      team.Player1,
				"player3_id":      team.Player2,
				"player4_id":      team.Player3,
				"player5_id":      team.Player4,
				"subplayer_id":    team.SubPlayer,
				"admin_player_id": team.Admin,
			},
		},
		Upsert:    false,
		Remove:    false,
		ReturnNew: true,
	}

	info, err := collection.Find(bson.M{"_id": teamID}).Apply(change, nil)

	if err == mgo.ErrNotFound {
		team.TeamID = teamID
		err := collection.Insert(team)
		if err != nil {
			log.Fatalln("Err inserting team, err: ", err)
		}
		t.updatePlayers()
		getTeamLogo(team.Logo, team.Name)
	}

	if info.Updated == 1 {
		t.updatePlayers()
	}

	// count, err := collection.Count()
	// if err != nil {
	// 	log.Fatalln("unable to open collection team, err:", err)
	// }
	//
	// if count == 0 {
	// 	for _, i := range t.JSON.Result.Teams {
	// 		err = collection.Insert(i)
	// 		if err != nil {
	// 			log.Fatalln("unable insert team into mgo, err:", err)
	// 		}
	// 		getTeamLogo(i.Logo, i.Name)
	// 	}
	// } else {
	// 	for _, i := range t.JSON.Result.Teams {
	// 		count, err = collection.Find(bson.M{"team_id": i.TeamID}).Count()
	// 		if err != nil {
	// 			log.Fatalln("unable to find total count of team_id, err: ", err)
	// 		}
	// 		if count == 0 {
	// 			collection.Insert(i)
	// 			getTeamLogo(i.TeamID, i.Name)
	// 		}
	// 		//TODO: Update players
	// 	}
	// }
}

func (t *Team) updatePlayers() {
	//TODO: maybe use goroutine to update player details
	team := t.JSON.Result.Teams[0]
	player := GetPlayer(t.DB)
	player.getPlayer(team.Player0)
	player.getPlayer(team.Player1)
	player.getPlayer(team.Player2)
	player.getPlayer(team.Player3)
	player.getPlayer(team.Player4)
	player.getPlayer(team.SubPlayer)
	player.getPlayer(team.Admin)
}
