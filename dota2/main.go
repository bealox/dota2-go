package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gopkg.in/mgo.v2"
)

var steamDetails = steamKey{
	languages: []string{"en-us", "zh-cn"},
}

func init() {
	steamDetails.key = os.Getenv("STEAM_API")
}

type steamKey struct {
	key       string
	languages []string
}

//API interface, anyy API need to inherits it.
type API interface {
	getURLs() map[string]string
	getJSON() error
	run()
}

func main() {

	var stopLock sync.Mutex
	stop := false
	signalChan := make(chan os.Signal, 1)

	session := connectDB()
	db := session.DB("dota2")
	heroController := GetTeam(db)
	// heroController.TeamID = 350190
	heroController.getTeam(350190)
	// p := GetPlayer(db)
	// p.getPlayer(43276219) //envy
	// p.getPlayer(89871557) //envy

	go func() {
		<-signalChan
		stopLock.Lock()
		stop = true
		stopLock.Unlock()
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		if stop {
			break
		}
	}
	defer session.Close()
	log.Println("killed it")

}

func connectDB() *mgo.Session {
	conneciton, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalln("unable to connect go mango db - ", err)
	}

	return conneciton
}
