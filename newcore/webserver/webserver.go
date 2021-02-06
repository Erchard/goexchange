package webserver

import (
	"../stackserver"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var dataChannel chan<- interface{}

func StartServer(stackChannel chan<- interface{}) {
	dataChannel = stackChannel

	r := mux.NewRouter()
	r.HandleFunc("/pair", addPair).Methods("POST")
	r.HandleFunc("/order", addOrder).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", r))

}

func addOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order stackserver.OrderDTO
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		fmt.Printf(err.Error())
	}
	dataChannel <- order

}

func addPair(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var pair stackserver.PairDTO
	err := json.NewDecoder(r.Body).Decode(&pair)
	if err != nil {
		fmt.Printf(err.Error())
	}
	dataChannel <- pair

}
