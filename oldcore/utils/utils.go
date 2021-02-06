package utils

import (
	"../core"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	fmt.Println("Init")
	rand.Seed(time.Now().UnixNano())
}

var idCounter uint64 = 1

func CreateRandomOrder(market *core.Market, pairObject *core.Pair) []core.Event {

	events := market.AddNewOrder(idCounter,
		"BTC/USD",
		rand.Float32() < 0.5,
		rand.Float64()*10.0,
		rand.Float64()*100000.0)

	idCounter++

	for _, e := range events {
		fmt.Printf(" %x \n", e, e.Order)
	}

	return events
}

/*

func PrintOrdersPrice(pair *core.Pair) {

	selLen := len(pair.sellStack)
	buyLen := len(pair.buyStack)
	closeLen := len(pair.closeStack)
	fmt.Println("\nSell")
	for i := selLen - 1; i >= 0; i-- {
		fmt.Println(pair.sellStack[i])
	}
	fmt.Println("\nBuy")
	for i := 0; i < buyLen; i++ {
		fmt.Println(pair.buyStack[i])
	}
	fmt.Println("\nClose")
	for i := 0; i < closeLen; i++ {
		fmt.Println(pair.closeStack[i])
	}

}


*/
