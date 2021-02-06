package stackserver

import (
	"../reactor"
	"fmt"
)

type PairDTO struct {
	Currency1 string `json:"currency1"`
	Currency2 string `json:"currency2"`
}

type OrderDTO struct {
	Id        uint64  `json:"id"`
	PairName  string  `json:"pairName"`
	IsGreen   bool    `json:"isGreen"`
	Currency1 float64 `json:"currency1"`
	Currency2 float64 `json:"currency2"`
}

var inChannel <-chan interface{}
var outChannel chan<- reactor.Event

var market *reactor.Market

func StartServer(inData <-chan interface{}, outData chan<- reactor.Event) {
	market = reactor.CreateMarket()
	inChannel = inData
	outChannel = outData

	for {

		i := <-inChannel

		switch v := i.(type) {
		case OrderDTO:

			events := market.AddNewOrder(v.Id, v.PairName, v.IsGreen, v.Currency1, v.Currency2)

			for _, e := range events {
				outChannel <- e
			}

		case PairDTO:

			p, err := market.AddPair(v.Currency1, v.Currency2)
			if !err {
				fmt.Println(p)
			}

		default:
			fmt.Println("Type %T not found \n", v)
		}

	}

}
