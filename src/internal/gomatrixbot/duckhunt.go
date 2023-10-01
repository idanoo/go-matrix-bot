package gomatrixbot

import (
	"log"
	"math/rand"
	"time"
)

type duckHunt struct {
	rooms map[string]int64 // [roomID]nextDuckTimestamp
}

var (
	startHuntText = []string{
		"Ducks have been known to hide around this area... Keep an eye out..",
		"IT'S DUCKHUNTING SEASON!",
		"Ducks have been spotted nearby..",
	}

	endHuntText = []string{
		"The ducks have gone on holiday to get away from this chaos..",
		"The ducks are taking a rest now. You should too..",
	}

	alreadyHuntText = []string{
		"Ducks are already out there..",
	}

	noHuntText = []string{
		"There is no duck hunt to stop... >_>",
	}

	duckTail = []string{
		"・゜゜・。。・゜゜",
	}

	duckText = []string{
		"\\_o< ",
		"\\_O< ",
		"\\_0< ",
		"\\_\u00f6< ",
		"\\_\u00f8< ",
		"\\_\u00f3< ",
	}

	duckNoise = []string{
		"QUACK!",
		"FLAP FLAP!",
		"quack!",
	}
)

func startHunt() {

}

func stopHunt() {

}

func shoot() {

}

func (mtrx *MtrxClient) initDuckHunt() {
	rooms := mtrx.getDuckHunt()
	for roomName, _ := range rooms {
		_, err := mtrx.c.JoinRoom(roomName, "", nil)
		if err != nil {
			log.Print(err)
		}
	}
}

func generateDuck() string {
	duck := ""
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	randInt := r1.Intn(3)
	for i := 0; i < randInt; i++ {
		duck = duck + duckTail[0]
	}

	body := duckText[rand.Intn(len(duckText))]
	noise := duckNoise[rand.Intn(len(duckNoise))]

	return duck + " " + body + noise
}
