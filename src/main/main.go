package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/fernanlukban/elise"
	"github.com/yuhanfang/riot/apiclient"
	"github.com/yuhanfang/riot/constants/champion"
	"github.com/yuhanfang/riot/constants/lane"
	"github.com/yuhanfang/riot/constants/region"
	"github.com/yuhanfang/riot/ratelimit"
)

type championStatsMap map[champion.Champion]map[lane.Lane]int

var mutex = &sync.Mutex{}
var writeMutex = &sync.Mutex{}
var matchWg sync.WaitGroup
var playerWg sync.WaitGroup
var champMap = make(championStatsMap)
var sem = make(chan int, 1)
var playerCount = 0

const MAX_NUM = 1000

var gameCount = 0

func (cm *championStatsMap) add(lane lane.Lane, champ champion.Champion) {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	champMap[champ][lane]++
}

func getRoleAndChampionInMatch(m *apiclient.MatchReference) {
	matchWg.Add(1)
	defer mutex.Unlock()
	mutex.Lock()
	gameCount++
	fmt.Println(m.Lane)
	fmt.Println(m.Champion)
	fmt.Println(m.GameID)
	champMap.add(m.Lane, m.Champion)
	// fmt.Println()
	// fmt.Println(*m)
}

func walkPlayerMatches(matches *apiclient.Matchlist, matchList chan int64) {
	fmt.Println("Walking player matches")
	sem <- 1
	if playerCount == MAX_NUM {
		fmt.Println("REACHED 100 PLAYERS")
		<-sem
		return
	}
	fmt.Println("Incrementing", playerCount)
	playerCount++
	<-sem
	fmt.Println("Finished incrementing", playerCount)
	for _, match := range matches.Matches {
		match := match
		go getRoleAndChampionInMatch(&match)
		matchList <- match.GameID
		fmt.Println("Adding", match.GameID)
	}
}

func processGameID(client apiclient.Client, ctx *context.Context, reg region.Region, matchList chan int64) {
	for id := range matchList {
		match, _ := client.GetMatch(*ctx, reg, id)
		for _, participantIdentity := range match.ParticipantIdentities {
			accountID := participantIdentity.Player.AccountID
			matches, _ := client.GetMatchlist(*ctx, reg, accountID, nil)
			go walkPlayerMatches(matches, matchList)
			playerWg.Done()
		}
	}
}

func main() {
	key := os.Getenv("RIOT_APIKEY")
	httpClient := http.DefaultClient
	ctx := context.Background()
	limiter := ratelimit.NewLimiter()
	const reg = region.NA1
	client := apiclient.New(key, httpClient, limiter)
	pw := elise.NewPlayerWalker(client)
	gw := elise.NewGameWalker(client)
	op, _ := client.GetBySummonerName(ctx, reg, "opsadboys")
	gameSet := elise.NewGameSet()
	gameIDList, _ := pw.GetGameIDList(ctx, reg, op.AccountID, nil)
	gameIDQueue := make([]int64, 10)
	for _, gameID := range gameIDList {
		gameIDQueue = append(gameIDQueue, gameID)
		fmt.Print("HERE ")
		fmt.Println(gameID)
	}
	for gameSet.Count() < 1000 {
		currentGameID := gameIDQueue[gameSet.Count()]
		gameSet.Add(currentGameID)
		fmt.Println(currentGameID)
		accountIDList, _ := gw.GetAccountIDList(ctx, reg, currentGameID)
		for _, accountID := range accountIDList {
			gameIDList, _ := pw.GetGameIDList(ctx, reg, accountID, nil)
			for _, newGameID := range gameIDList {
				gameIDQueue = append(gameIDQueue, newGameID)
				fmt.Print("Here ")
				fmt.Println(newGameID)
			}
		}
	}
}

func op() {
	key := os.Getenv("RIOT_APIKEY")
	httpClient := http.DefaultClient
	ctx := context.Background()
	limiter := ratelimit.NewLimiter()
	const reg = region.NA1
	client := apiclient.New(key, httpClient, limiter)
	//pw := elise.NewPlayerWalker(client)
	// pw.GetGameIDList(12)

	for _, champion := range champion.All() {
		champMap[champion] = make(map[lane.Lane]int)
	}

	op, err := client.GetBySummonerName(ctx, reg, "opsadboys")
	matches, err := client.GetMatchlist(ctx, reg, op.AccountID, nil)

	if err != nil {
		fmt.Println("ERROR")
		return
	}
	matchList := make(chan int64, 4)

	playerWg.Add(MAX_NUM)

	go walkPlayerMatches(matches, matchList)
	go processGameID(client, &ctx, reg, matchList)

	matchWg.Wait()
	playerWg.Wait()

	var total = 0
	for _, champion := range champion.All() {
		fmt.Println(champion, lane.Middle, champMap[champion][lane.Middle])
		fmt.Println(champion, lane.Top, champMap[champion][lane.Top])
		fmt.Println(champion, lane.Bottom, champMap[champion][lane.Bottom])
		fmt.Println(champion, lane.Jungle, champMap[champion][lane.Jungle])
		total += champMap[champion][lane.Middle]
		total += champMap[champion][lane.Top]
		total += champMap[champion][lane.Bottom]
		total += champMap[champion][lane.Jungle]
	}

	fmt.Println(gameCount, total)
}
