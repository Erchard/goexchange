package reactor

import (
	"fmt"
	"math"
	"time"
)

type Market struct {
	pairMap     map[string]*Pair
	orderMap    map[uint64]*Order
	lastEventId uint64
	lastEvents  []Event
	fraction    float64
}

type Pair struct {
	currency1   string `json:"currency1"`
	currency2   string `json:"currency2"`
	buyStack    []*Order
	sellStack   []*Order
	curr1volume uint64
	curr2volume uint64
	//lastPrice   uint64
}

type Money struct {
	Currency string `json:"currency"`
	Amount   uint64 `json:"amount"`
}

type Order struct {
	Id            uint64 `json:"id"`
	PairName      string `json:"pair"`
	pair          *Pair
	IsGreen       bool   `json:"isGreen"`
	Price         uint64 `json:"price"`
	IsMarketPrice bool   `json:"isMarketPrice"`
	Want          Money  `json:"want"`
	Supply        Money  `json:"supply"`
	Received      Money  `json:"received"`
	IsClose       bool   `json:"isClose"`
}

type Swap struct {
	pair       *Pair
	Green      *Order `json:"green"`
	Red        *Order `json:"red"`
	Price      uint64 `json:"price"`
	Money1     uint64 `json:"money1"`
	Money2     uint64 `json:"money2"`
	Remainder1 uint64 `json:"remainder1"`
	Remainder2 uint64 `json:"remainder2"`
}

var market Market

type EventData interface {
}

type EventType int

const (
	Create = iota
	SwapOrder
	Cancel
	Error
)

type Event struct {
	Id        uint64    `json:"id"`
	Time      int64     `json:"time"`
	EventType EventType `json:"type"`
	Order     *Order    `json:"order"`
	Swap      *Swap     `json:"swap"`
}

func CreateMarket() *Market {
	market = Market{
		pairMap:     make(map[string]*Pair),
		orderMap:    make(map[uint64]*Order),
		lastEventId: 0,
		lastEvents:  make([]Event, 0),
		fraction:    10000,
	}
	return &market
}

func (m *Market) AddPair(currency1 string, currency2 string) (*Pair, bool) {
	pairName := currency1 + "/" + currency2
	_, exists := m.pairMap[pairName]
	if !exists {
		m.pairMap[pairName] = preparePair(currency1, currency2)
		//m.pairMap[pairName] = preparePair(currency1, currency2, uint64(math.Round(lastPrice*m.fraction*m.fraction)))
		return m.pairMap[pairName], false
	} else {
		fmt.Printf("Pair %s exists\n", pairName)
		return nil, true
	}
}

func (m *Market) newOrderEvent(order *Order) {
	m.lastEventId++
	event := Event{
		Id:        m.lastEventId,
		Time:      time.Now().UnixNano(),
		EventType: Create,
		Order:     order,
	}
	m.lastEvents = append(m.lastEvents, event)
}

func preparePair(currency1 string, currency2 string) *Pair {
	pair := Pair{
		currency1:   currency1,
		currency2:   currency2,
		buyStack:    make([]*Order, 0),
		sellStack:   make([]*Order, 0),
		curr1volume: 0,
		curr2volume: 0,
		//lastPrice:   lastPrice,
	}

	return &pair
}

func (m *Market) AddNewOrder(id uint64, pairName string, isGreen bool, currency1 float64, currency2 float64) []Event {
	m.lastEvents = m.lastEvents[:0]
	order, err := prepareOrder(id, pairName, isGreen, currency1, currency2)
	if !err {
		market.orderMap[id] = order

		order.addToStack()
	} else {
		m.lastEventId++
		event := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Error,
			Order:     order,
		}
		m.lastEvents = append(m.lastEvents, event)
	}
	return m.lastEvents
}

func (m *Market) CancelOrder(id uint64) []Event {
	m.lastEvents = m.lastEvents[:0]
	order, exists := m.orderMap[id]
	m.lastEventId++
	event := Event{
		Id:   m.lastEventId,
		Time: time.Now().UnixNano(),
	}
	if exists {
		order.close()
		event.EventType = Cancel
		event.Order = order
	} else {
		event.EventType = Error
	}
	m.lastEvents = append(m.lastEvents, event)
	return m.lastEvents
}

func prepareOrder(id uint64, pairName string, isGreen bool, currency1 float64, currency2 float64) (*Order, bool) {

	_, exists := market.orderMap[id]
	if exists {
		fmt.Printf("Order with id %d exists \n", id)
		return nil, true
	}

	pair, exists := market.pairMap[pairName]
	if !exists {
		fmt.Printf("Pair %s not found \n", pairName)
		return nil, true
	}

	price, isMarketPrice, err := calcPrice(isGreen, currency1, currency2)
	if err {
		fmt.Printf("Error calc price. \n isGreen: %x \t Currency 1: %d \t Currency 2: %d \n",
			isGreen, currency1, currency2)
		return nil, true
	}

	order := Order{
		Id:            id,
		PairName:      pairName,
		pair:          pair,
		IsGreen:       isGreen,
		Price:         price,
		IsMarketPrice: isMarketPrice,
		IsClose:       false,
	}

	var wantAmount uint64

	var supplyAmount uint64
	if isGreen {
		supplyAmount = uint64(math.Round(currency2 * market.fraction))
		if isMarketPrice {
			wantAmount = math.MaxUint64
		} else {
			wantAmount = uint64(math.Round(currency1 * market.fraction))
		}

		order.Want = Money{
			Currency: pair.currency1,
			Amount:   wantAmount,
		}

		order.Supply = Money{
			Currency: pair.currency2,
			Amount:   supplyAmount,
		}
		order.Received = Money{
			Currency: pair.currency1,
			Amount:   0,
		}
	} else {
		supplyAmount = uint64(math.Round(currency1 * market.fraction))
		if isMarketPrice {
			wantAmount = 0
		} else {
			wantAmount = uint64(math.Round(currency2 * market.fraction))
		}
		order.Want = Money{
			Currency: pair.currency2,
			Amount:   wantAmount,
		}

		order.Supply = Money{
			Currency: pair.currency1,
			Amount:   supplyAmount,
		}
		order.Received = Money{
			Currency: pair.currency2,
			Amount:   0,
		}
	}

	return &order, false
}

func calcPrice(isGreen bool, currency1 float64, currency2 float64) (price uint64, isMarketPrice bool, err bool) {
	if currency1 != 0 && currency2 != 0 {
		price = uint64(math.Round(currency2/currency1) * market.fraction * market.fraction)
		return price, false, false
	} else if currency1 == 0 && currency2 == 0 {
		return 0, true, true
	} else if currency1 == 0 && isGreen {
		price = math.MaxUint64
		return price, true, false
	} else if currency2 == 0 && !isGreen {
		price = 0
		return price, true, false
	} else {
		return 0, true, true
	}
}

func (o *Order) give(amount uint64) uint64 {
	var give uint64 = 0
	if !o.IsMarketPrice {
		if o.IsGreen {
			give = uint64(math.Round((float64(amount) * (float64(o.Price) / market.fraction)) /
				market.fraction))
		} else {
			give = uint64(math.Round(((float64(amount) * market.fraction) / float64(o.Price)) *
				market.fraction))
		}
	}
	return give
}

func (o *Order) addMoney(amount uint64, give uint64) {

	if !o.IsMarketPrice {
		if o.Supply.Amount >= give && o.Want.Amount >= amount {
			o.Supply.Amount -= give
			o.Want.Amount -= amount
			o.Received.Amount += amount
		} else {
			give = 0
			fmt.Printf("Not enough money.\n Supply: %d \t Give: %d \t Want: %d \t Amount: %d \n",
				o.Supply.Amount, give, o.Want.Amount, amount)
		}
	}
}

func (o *Order) takeMoney(amount uint64) {

}

func (o *Order) addToStack() {
	if o.IsGreen {
		o.pair.curr2volume += o.Supply.Amount
		o.addToBuyStack()
	} else {
		o.pair.curr1volume += o.Supply.Amount
		o.addToSellStack()
	}
	market.newOrderEvent(o)
	o.pair.swap()
}

func (o *Order) close() {
	o.IsClose = true
	if o.IsGreen {
		o.pair.curr1volume -= o.Received.Amount
		o.pair.curr2volume -= o.Supply.Amount
		o.removeFromBuyStack()
	} else {
		o.pair.curr1volume -= o.Supply.Amount
		o.pair.curr2volume -= o.Received.Amount
		o.removeFromSellStack()
	}
}

func (o *Order) addToSellStack() {
	stackLen := len(o.pair.sellStack)
	last := stackLen - 1
	if stackLen < 1 || o.Price >= o.pair.sellStack[last].Price {
		o.pair.sellStack = append(o.pair.sellStack, o)
	} else if o.Price < o.pair.sellStack[0].Price {
		o.pair.sellStack = append(o.pair.sellStack, o.pair.sellStack[last])
		copy(o.pair.sellStack[1:], o.pair.sellStack[:last])
		o.pair.sellStack[0] = o
	} else {
		for i := 1; i < stackLen; i++ {
			if o.Price >= o.pair.sellStack[i-1].Price && o.Price < o.pair.sellStack[i].Price {
				o.pair.sellStack = append(o.pair.sellStack, o.pair.sellStack[last])
				copy(o.pair.sellStack[i+1:], o.pair.sellStack[i:last])
				o.pair.sellStack[i] = o
				break
			}
		}
	}
	o.pair.curr1volume += o.Supply.Amount
}

func (o *Order) addToBuyStack() {
	stackLen := len(o.pair.buyStack)
	last := stackLen - 1
	if stackLen < 1 || o.Price <= o.pair.buyStack[last].Price {
		o.pair.buyStack = append(o.pair.buyStack, o)
	} else if o.Price > o.pair.buyStack[0].Price {
		o.pair.buyStack = append(o.pair.buyStack, o.pair.buyStack[last])
		copy(o.pair.buyStack[1:], o.pair.buyStack[:last])
		o.pair.buyStack[0] = o
	} else {
		for i := 1; i < stackLen; i++ {
			if o.Price <= o.pair.buyStack[i-1].Price && o.Price > o.pair.buyStack[i].Price {
				o.pair.buyStack = append(o.pair.buyStack, o.pair.buyStack[last])
				copy(o.pair.buyStack[i+1:], o.pair.buyStack[i:last])
				o.pair.buyStack[i] = o
				break
			}
		}
	}
	o.pair.curr2volume += o.Supply.Amount
}

func (o *Order) removeFromSellStack() {
	stackLen := len(o.pair.sellStack)
	if stackLen < 2 {
		o.pair.sellStack = o.pair.sellStack[:0]
	} else {
		for i, order := range o.pair.sellStack {
			if order == o {
				copy(o.pair.sellStack[i:], o.pair.sellStack[i+1:])
				o.pair.sellStack = o.pair.sellStack[:stackLen-1]
				break
			}
		}
	}
}

func (o *Order) removeFromBuyStack() {
	stackLen := len(o.pair.buyStack)
	if stackLen < 2 {
		o.pair.buyStack = o.pair.buyStack[:0]
	} else {
		for i, order := range o.pair.buyStack {
			if order == o {
				copy(o.pair.buyStack[i:], o.pair.buyStack[i+1:])
				o.pair.buyStack = o.pair.buyStack[:stackLen-1]
				break
			}
		}
	}
}

func (p *Pair) swap() {
	if len(p.buyStack) > 0 && len(p.sellStack) > 0 && p.buyStack[0].Price >= p.sellStack[0].Price {

		swap := Swap{
			pair:  p,
			Green: p.buyStack[0],
			Red:   p.sellStack[0],
		}

		if swap.Red.IsMarketPrice {
			if swap.Red.Supply.Amount > swap.Green.Want.Amount {
				swap.case5()
			} else if swap.Red.Supply.Amount < swap.Green.Want.Amount {
				swap.case8()
			} else if swap.Red.Supply.Amount == swap.Green.Want.Amount {
				swap.case9()
			} else {
				fmt.Println("Case not found!!")
			}
		} else if swap.Green.IsMarketPrice {
			if swap.Green.Supply.Amount > swap.Red.Want.Amount {
				swap.case2()
			} else if swap.Green.Supply.Amount < swap.Red.Want.Amount {
				swap.case7()
			} else if swap.Green.Supply.Amount == swap.Red.Want.Amount {
				swap.case10()
			} else {
				fmt.Println("Case not found!!")
			}
		} else if swap.Red.Supply.Amount > swap.Green.Want.Amount {
			if swap.Green.Supply.Amount > swap.Red.Want.Amount {
				swap.case1()
			} else if swap.Green.Supply.Amount < swap.Red.Want.Amount {
				swap.case6()
			} else if swap.Green.Supply.Amount == swap.Red.Want.Amount {
				swap.case11()
			} else {
				fmt.Println("Case not found!!")
			}
		} else if swap.Red.Supply == swap.Green.Want {
			if swap.Green.Supply.Amount > swap.Red.Want.Amount {
				swap.case4()
			} else if swap.Green.Supply.Amount == swap.Red.Want.Amount {
				swap.case12()
			} else {
				fmt.Println("Case not found!!")
			}
		} else if swap.Red.Supply.Amount < swap.Green.Want.Amount {
			if swap.Green.Supply.Amount > swap.Red.Want.Amount {
				swap.case3()
			} else {
				fmt.Println("Case not found!!")
			}
		} else {
			fmt.Println("Case not found!!")
		}

		market.lastEventId++

		event := Event{
			Id:        market.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: SwapOrder,
			Swap:      &swap,
		}

		market.lastEvents = append(market.lastEvents, event)
		//p.lastPrice = swap.Price

		fmt.Printf("Swap: %+v \n", swap)

		p.swap()
	}
}

func (s *Swap) case1() {
	fmt.Println("Red.Supply > Green.Want")
	fmt.Println("Green.Supply > Red.Want")

	s.Money1 = s.Green.Want.Amount
	s.Money2 = s.Red.Want.Amount
	s.Price = uint64(math.Round(float64(s.Money2)/float64(s.Money1)) * market.fraction * market.fraction)

	s.Remainder1 = s.Red.Supply.Amount - s.Green.Want.Amount
	s.Remainder2 = s.Green.Supply.Amount - s.Red.Want.Amount

	s.Red.Supply.Amount = 0
	s.Red.Received.Amount += s.Money2
	s.Red.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Red.close()
	s.Green.close()
	s.pair.curr1volume -= s.Remainder1
	s.pair.curr2volume -= s.Remainder2

}

func (s *Swap) case2() {
	fmt.Println("Green.Supply > Red.Want")
	fmt.Println("Green.IsMarketPrice")

	s.Price = s.Red.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Red.Want.Amount

	s.Red.Supply.Amount -= s.Money1
	s.Green.Received.Amount += s.Money1

	s.Green.Supply.Amount -= s.Money2
	s.Red.Want.Amount -= s.Money2
	s.Red.Received.Amount += s.Money2

	s.Red.close()

}

func (s *Swap) case3() {
	fmt.Println("Green.Supply > Red.Want")
	fmt.Println("Red.Supply < Green.Want")

	s.Price = s.Red.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Red.Want.Amount

	s.Remainder2 = s.Green.give(s.Money2) - s.Red.Want.Amount

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount -= s.Money1

	s.Green.Supply.Amount -= (s.Money2 + s.Remainder2)
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
	s.pair.curr2volume -= s.Remainder2

}

func (s *Swap) case4() {
	fmt.Println("Green.Supply > Red.Want")
	fmt.Println("Red.Supply = Green.Want")

	s.Price = s.Red.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Red.Want.Amount

	s.Remainder2 = s.Green.Supply.Amount - s.Red.Want.Amount

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
	s.Green.close()
	s.pair.curr2volume -= s.Remainder2
}

func (s *Swap) case5() {
	fmt.Println("Red.Supply > Green.Want")
	fmt.Println("Red.IsMarketPrice")

	s.Price = s.Green.Price
	s.Money1 = s.Green.Want.Amount
	s.Money2 = s.Green.Supply.Amount

	s.Red.Supply.Amount -= s.Money1
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Red.Received.Amount += s.Money2
	s.Green.Supply.Amount = 0

	s.Green.close()
}

func (s *Swap) case6() {
	fmt.Println("Green.Supply < Red.Want")
	fmt.Println("Red.Supply > Green.Want")

	s.Price = s.Green.Price
	s.Money1 = s.Green.Want.Amount
	s.Money2 = s.Green.Supply.Amount

	s.Remainder1 = s.Red.give(s.Money2) - s.Green.Want.Amount

	s.Green.Supply.Amount = 0
	s.Red.Received.Amount += s.Money2
	s.Red.Want.Amount -= s.Money2

	s.Red.Supply.Amount -= (s.Money1 + s.Remainder1)
	s.Green.Want.Amount = 0
	s.Green.Received.Amount += s.Money1

	s.Green.close()
	s.pair.curr1volume -= s.Remainder1
}

func (s *Swap) case7() {
	fmt.Println("Green.Supply < Red.Want")
	fmt.Println("Green.IsMarketPrice")

	s.Price = s.Red.Price
	s.Money2 = s.Green.Supply.Amount
	s.Money1 = s.Red.give(s.Money2)

	s.Red.Supply.Amount -= s.Money1
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Red.Want.Amount -= s.Money2
	s.Red.Received.Amount += s.Money2

	s.Green.close()

}

func (s *Swap) case8() {
	fmt.Println("Red.Supply < Green.Want")
	fmt.Println("Red.IsMarketPrice")

	s.Price = s.Green.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Green.give(s.Money1)

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount -= s.Money1

	s.Green.Supply.Amount -= s.Money2
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
}

func (s *Swap) case9() {
	fmt.Println("Red.Supply = Green.Want")
	fmt.Println("Red.IsMarketPrice")

	s.Price = s.Green.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Green.Supply.Amount

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
	s.Green.close()
}

func (s *Swap) case10() {
	fmt.Println("Green.Supply = red.Want")
	fmt.Println("Green.IsMarketPrice")

	s.Price = s.Red.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Green.Supply.Amount

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
	s.Green.close()
}

func (s *Swap) case11() {
	fmt.Println("Red.Supply > Green.Want")
	fmt.Println("Green.Supply = Red.Want")

	s.Price = s.Green.Price
	s.Money1 = s.Green.Want.Amount
	s.Money2 = s.Green.Supply.Amount

	s.Remainder1 = s.Red.Supply.Amount - s.Green.Want.Amount

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
	s.Green.close()
	s.pair.curr1volume -= s.Remainder1
}

func (s *Swap) case12() {
	fmt.Println("Green.Supply = Red.Want")
	fmt.Println("Red.Supply = Green.Want")

	s.Price = s.Red.Price
	s.Money1 = s.Red.Supply.Amount
	s.Money2 = s.Red.Want.Amount

	s.Red.Supply.Amount = 0
	s.Green.Received.Amount += s.Money1
	s.Green.Want.Amount = 0

	s.Green.Supply.Amount = 0
	s.Red.Want.Amount = 0
	s.Red.Received.Amount += s.Money2

	s.Red.close()
	s.Green.close()
}
