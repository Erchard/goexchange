package main

import (
	"./reactor"
	"./stackserver"
	"./webserver"
	"encoding/json"
	"fmt"
	"log"
)

var market *reactor.Market

func main() {
	fmt.Println("Start")
	//testMarket()
	testChannels()
}

func testChannels() {
	var ch1 = make(chan interface{}, 10000)
	var ch2 = make(chan reactor.Event, 10000)

	go stackserver.StartServer(ch1, ch2)
	go webserver.StartServer(ch1)

	go func() {
		m, err := json.Marshal(<-ch2)
		if err == nil {
			fmt.Println(string(m))
		}
	}()

	var input string
	fmt.Scanln(&input)
}

func testMarket() {
	market = reactor.CreateMarket()
	var eventList []reactor.Event

	market.AddPair("BTC", "USD")
	eventList = market.AddNewOrder(1, "BTC/USD", false, 1.9, 65000.01)
	json.NewEncoder(log.Writer()).Encode(eventList)
	eventList = market.AddNewOrder(2, "BTC/USD", false, 2.31, 65000)
	json.NewEncoder(log.Writer()).Encode(eventList)
	eventList = market.CancelOrder(1)
	json.NewEncoder(log.Writer()).Encode(eventList)
	eventList = market.AddNewOrder(3, "BTC/USD", true, 0, 100000.12)
	json.NewEncoder(log.Writer()).Encode(eventList)
	eventList = market.AddNewOrder(4, "BTC/USD", false, 1.9, 65000.01)
	json.NewEncoder(log.Writer()).Encode(eventList)
}
