package main

import (
	"./core"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type CurrencyDTO struct {
	Name    string `json:"name"`
	Decimal uint8  `json:"decimal"`
}

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

var market *core.Market

func init() {
	// Включаем номера строк в логах
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	testMarket()
}

/*
func startServer() {

	market = core.CreateMarket()
	r := mux.NewRouter()
	r.HandleFunc("/currency", addCurrency).Methods("POST")
	r.HandleFunc("/pair", addPair).Methods("POST")
	r.HandleFunc("/order", addOrder).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", r))


}
*/
func addOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order OrderDTO
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		fmt.Printf(err.Error())
	}
	eventList := market.AddNewOrder(
		order.Id,
		order.PairName,
		order.IsGreen,
		order.Currency1,
		order.Currency2,
	)
	json.NewEncoder(w).Encode(eventList)
}

func addPair(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var pair PairDTO
	err := json.NewDecoder(r.Body).Decode(&pair)
	if err != nil {
		fmt.Printf(err.Error())
	}
	market.AddPair(pair.Currency1, pair.Currency2)
	json.NewEncoder(w).Encode(pair)
}

func addCurrency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var currency CurrencyDTO
	err := json.NewDecoder(r.Body).Decode(&currency)
	if err != nil {
		fmt.Printf(err.Error())
	}
	market.AddCurrency(currency.Name, currency.Decimal)
	json.NewEncoder(w).Encode(currency)
}

func testMarket() {
	market = core.CreateMarket()
	market.AddCurrency("BTC", 4)
	market.AddCurrency("USD", 2)
	market.AddPair("BTC", "USD")
	market.AddNewOrder(1, "BTC/USD", false, 1.9, 65000.01)
	market.AddNewOrder(2, "BTC/USD", false, 2.31, 65000)
	eventList := market.AddNewOrder(3, "BTC/USD", true, 0, 100000.12)

	json.NewEncoder(log.Writer()).Encode(eventList)
}

//func handler(w http.ResponseWriter, r *http.Request){
//	fmt.Fprintf(w, "Hi there, I love %s!",r.URL.Path[1:])
//}
//
//func main(){
//	http.HandleFunc("/",handler)
//	log.Fatal(http.ListenAndServe(":8080",nil))
//}

/*

func main() {

	market := core.CreateMarket()

	market.AddCurrency("BTC", 4)
	market.AddCurrency("USD", 2)
	pairObject,_ := market.AddPair("BTC", "USD")

	start := time.Now()
	for i := 0; i < 100; i++ {
		utils.CreateRandomOrder(market, pairObject)
	}
	end := time.Now()

	//utils.PrintOrdersPrice(pairObject)
	fmt.Println(start)
	fmt.Println(end)

}

*/

/*
func testClose() {
	func main() {

	order1 := core.Order{
		Id:      1,
		IsGreen: true,
		Curr1:   19000,   // 2BTC
		Curr2:   6500000, // 65'000 USD
		IsClose: false,
	}
	prepareOrder(&order1, pairObject)

	order2 := Order{
		id:      2,
		isGreen: false,
		curr1:   20000,   // 1.9BTC
		curr2:   6600000, // 60'000 USD
		isClose: false,
	}
	prepareOrder(&order2, pairObject)

	addToStack(&order1)
	addToStack(&order2)

	printOrdersPrice(pairObject)

	closeOrders(pairObject)

	printOrdersPrice(pairObject)

*/
