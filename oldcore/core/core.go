package core

import (
	"fmt"
	"math"
	"time"
)

type Market struct {
	currencyMap map[string]*Currency
	pairMap     map[string]*Pair
	orderMap    map[uint64]*Order
	lastEventId uint64
	lastEvents  []Event
}

type Currency struct {
	name     string `json:"name"`
	fraction uint64
}

type Pair struct {
	currency1   *Currency `json:"currency1"`
	currency2   *Currency `json:"currency2"`
	buyStack    []Order
	sellStack   []Order
	fraction    uint64
	curr1volume uint64
	curr2volume uint64
	lastPrice   uint64
}

type Order struct {
	Id            uint64 `json:"id"`
	IsGreen       bool   `json:"isGreen"`
	Curr1         uint64 `json:"currency1"`
	Curr2         uint64 `json:"currency2"`
	Price         uint64 `json:"price"`
	Curr1volume   uint64 `json:"curr1volume"`
	Curr2volume   uint64 `json:"curr2volume"`
	IsClose       bool   `json:"isClose"`
	StockBonus    uint64 `json:"stockBonus"`
	IsMarketPrice bool   `json:"isMarketPrice"`
	PairName      string `json:"pair"`
	pair          *Pair
}

type EventType int

const (
	Create = iota
	Change
	Remove
	Error
)

type Event struct {
	Id        uint64    `json:"id"`
	Time      int64     `json:"time"`
	EventType EventType `json:"type"`
	Order     *Order    `json:"order"`
}

func CreateMarket() *Market {
	market := Market{
		currencyMap: make(map[string]*Currency),
		pairMap:     make(map[string]*Pair),
		orderMap:    make(map[uint64]*Order),
		lastEventId: 0,
		lastEvents:  make([]Event, 0),
	}
	return &market
}

func (m *Market) AddCurrency(name string, decimal uint8) (*Currency, bool) {

	if _, exists := m.currencyMap[name]; !exists {
		m.currencyMap[name] = &Currency{
			name:     name,
			fraction: uint64(math.Pow(10, float64(decimal))),
		}
	} else {
		fmt.Printf("Currency %s exists\n", name)
		return nil, false
	}
	return m.currencyMap[name], true
}

func (m *Market) AddPair(currency1 string, currency2 string) (*Pair, bool) {
	pairName := currency1 + "/" + currency2
	cur1, ok1 := m.currencyMap[currency1]
	cur2, ok2 := m.currencyMap[currency2]
	_, pOk := m.pairMap[pairName]
	if ok1 && ok2 && !pOk {
		m.pairMap[pairName] = preparePair(cur1, cur2)
	} else {
		if !ok1 {
			fmt.Printf("Currency %s not found\n", currency1)
		}
		if !ok2 {
			fmt.Printf("Currency %s not found\n", currency2)
		}
		if pOk {
			fmt.Printf("Pair %s exists\n", pairName)
		}
		return nil, false
	}
	return m.pairMap[pairName], true
}

func (m *Market) AddNewOrder(id uint64, pairName string, isGreen bool, currency1 float64, currency2 float64) []Event {
	m.lastEvents = m.lastEvents[:0]

	o, ok := m.orderMap[id]
	if ok {
		fmt.Printf("Order %d exists\n", id)
		m.lastEventId++
		ev := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Error,
			Order:     o,
		}
		m.lastEvents = append(m.lastEvents, ev)
		return m.lastEvents
	}

	pair, ok := m.pairMap[pairName]
	if !ok {
		fmt.Printf("Pair %s not found\n", pairName)
		m.lastEventId++
		ev := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Error,
			Order:     nil,
		}
		m.lastEvents = append(m.lastEvents, ev)
		return m.lastEvents
	}

	if currency1 == 0 && currency2 == 0 {
		fmt.Printf("amount = 0")
		m.lastEventId++
		ev := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Error,
			Order:     nil,
		}
		m.lastEvents = append(m.lastEvents, ev)
		return m.lastEvents
	}

	order := Order{
		Id:            id,
		IsGreen:       isGreen,
		Curr1:         uint64(currency1 * float64(pair.currency1.fraction)),
		Curr2:         uint64(currency2 * float64(pair.currency2.fraction)),
		IsClose:       false,
		StockBonus:    0,
		IsMarketPrice: ((currency1 == 0 && isGreen) || (currency2 == 0 && !isGreen)) && currency1 != currency2,
		pair:          pair,
	}
	order.StockBonus = 0

	if order.IsGreen {
		order.Curr1volume = 0
		order.Curr2volume = order.Curr2
		order.IsMarketPrice = order.Curr1 == 0
	} else {
		order.Curr1volume = order.Curr1
		order.Curr2volume = 0
		order.IsMarketPrice = order.Curr2 == 0
	}

	setPrice(&order)
	m.orderMap[id] = &order

	if order.IsMarketPrice {
		m.closeByMarket(&order)
	} else {
		m.addToStack(&order)
	}
	return m.lastEvents
}

func preparePair(currency1 *Currency, currency2 *Currency) *Pair {
	pair := Pair{
		currency1: currency1,
		currency2: currency2,
		buyStack:  make([]Order, 0),
		sellStack: make([]Order, 0),
		fraction:  currency1.fraction * currency2.fraction,
	}
	return &pair
}

func setPrice(order *Order) *Order {
	var c1 float64
	var c2 float64

	if order.Curr1 == 0 && order.Curr2 == 0 {
		return order
	}

	if order.IsGreen {
		if order.IsMarketPrice && order.Curr1 == 0 {
			order.Price = math.MaxUint64
			return order
		}

		c1 = float64(order.Curr1 - order.Curr1volume)
		c2 = float64(order.Curr2volume)
	} else {
		if order.IsMarketPrice && order.Curr2 == 0 {
			order.Price = 0
			return order
		}
		c1 = float64(order.Curr1volume)
		c2 = float64(order.Curr2 - order.Curr2volume)
	}

	floatprice := c2 * float64(order.pair.fraction) * float64(order.pair.currency2.fraction) / c1
	order.Price = uint64(floatprice)

	return order
}

func (m *Market) addToStack(order *Order) {

	if order.IsGreen {
		stackPointer := &order.pair.buyStack
		stack := *stackPointer

		last := len(stack) - 1
		if len(stack) < 1 || order.Price <= (stack)[last].Price {
			stack = append(stack, *order)
		} else if order.Price > stack[0].Price {
			stack = append(stack, stack[last])
			copy(stack[1:], stack[:last])
			stack[0] = *order
		} else {
			for i := 1; i < len(stack); i++ {
				if order.Price <= stack[i-1].Price && order.Price > stack[i].Price {
					stack = append(stack, stack[last])
					copy(stack[i+1:], stack[i:last])
					stack[i] = *order
					break
				}
			}
		}

		*stackPointer = stack

	} else {
		stackPointer := &order.pair.sellStack

		stack := *stackPointer

		last := len(stack) - 1
		if len(stack) < 1 || order.Price >= (stack)[last].Price {
			stack = append(stack, *order)
		} else if order.Price < stack[0].Price {
			stack = append(stack, stack[last])
			copy(stack[1:], stack[:last])
			stack[0] = *order
		} else {
			for i := 1; i < len(stack); i++ {
				if order.Price >= stack[i-1].Price && order.Price < stack[i].Price {
					stack = append(stack, stack[last])
					copy(stack[i+1:], stack[i:last])
					stack[i] = *order
					break
				}
			}
		}
		*stackPointer = stack
	}
	m.lastEventId++
	ev := Event{
		Id:        m.lastEventId,
		Time:      time.Now().UnixNano(),
		EventType: Create,
		Order:     order,
	}
	m.lastEvents = append(m.lastEvents, ev)

	m.closeOrders(order.pair)
	//return order.pair
}

func (m *Market) closeOrders(pair *Pair) {

	if len(pair.buyStack) < 1 || len(pair.sellStack) < 1 {
		return
	}

	orderToBuy := pair.buyStack[0]
	orderToSell := pair.sellStack[0]

	if orderToBuy.Price >= orderToSell.Price {
		fmt.Println("\nClose order")
		sellerWant := orderToSell.Curr2 - orderToSell.Curr2volume
		buyerWant := orderToBuy.Curr1 - orderToBuy.Curr1volume

		fmt.Println("Seller: ", sellerWant)
		fmt.Println("Buyer: ", buyerWant)

		if sellerWant <= orderToBuy.Curr2volume &&
			buyerWant >= orderToSell.Curr1volume {
			fmt.Println("Case 1")
			buyerGive := uint64(math.Round(float64(orderToSell.Curr1volume*orderToBuy.Price) /
				float64(pair.currency1.fraction) / float64(pair.currency1.fraction)))
			fmt.Println("Buyer Give: ", buyerGive)

			orderToBuy.Curr1volume += orderToSell.Curr1volume
			orderToSell.Curr1volume = 0

			if orderToBuy.IsMarketPrice {
				orderToBuy.Curr2volume -= sellerWant
				orderToSell.Curr2volume += sellerWant
			} else {
				orderToBuy.Curr2volume -= buyerGive
				orderToSell.Curr2volume += sellerWant
				orderToSell.StockBonus += (buyerGive - sellerWant)
			}

			fmt.Println("Stock Bonus:", orderToSell.StockBonus)

			if orderToBuy.Curr2volume == 0 {
				orderToBuy.IsClose = true
				if orderToBuy.IsMarketPrice {
					orderToBuy.Curr1volume += orderToBuy.StockBonus
					orderToBuy.StockBonus = 0
				}

				if len(pair.buyStack) > 1 {
					pair.buyStack = pair.buyStack[1:]
				} else {
					pair.buyStack = pair.buyStack[:0]
				}
			} else {
				setPrice(&orderToBuy)
				pair.buyStack[0] = orderToBuy
			}

			orderToSell.IsClose = true
			if orderToSell.IsMarketPrice {
				orderToSell.Curr2volume += orderToSell.StockBonus
				orderToSell.StockBonus = 0
			}

			if len(pair.sellStack) > 1 {
				pair.sellStack = pair.sellStack[1:]
			} else {
				pair.sellStack = pair.sellStack[:0]
			}

		} else if buyerWant <= orderToSell.Curr1volume &&
			sellerWant >= orderToBuy.Curr2volume {
			fmt.Println("Case 2")
			sellerGive := uint64(math.Round(float64(orderToBuy.Curr2volume*pair.currency1.fraction*pair.currency1.fraction) /
				float64(orderToSell.Price)))
			fmt.Println("Seller Give: ", sellerGive)

			orderToSell.Curr2volume += orderToBuy.Curr2volume
			orderToBuy.Curr2volume = 0

			orderToSell.Curr1volume -= sellerGive
			orderToBuy.Curr1volume += buyerWant
			orderToBuy.StockBonus += (sellerGive - buyerWant)
			fmt.Println("Stock Bonus: ", orderToBuy.StockBonus)

			if orderToSell.Curr1volume == 0 {
				orderToSell.IsClose = true
				if orderToSell.IsMarketPrice {
					orderToSell.Curr2volume += orderToSell.StockBonus
					orderToSell.StockBonus = 0
				}

				if len(pair.sellStack) > 1 {
					pair.sellStack = pair.sellStack[1:]
				} else {
					pair.sellStack = pair.sellStack[:0]
				}
			} else {
				setPrice(&orderToSell)
				pair.sellStack[0] = orderToSell
			}

			orderToBuy.IsClose = true
			if orderToBuy.IsMarketPrice {
				orderToBuy.Curr1volume += orderToBuy.StockBonus
				orderToBuy.StockBonus = 0
			}

			if len(pair.buyStack) > 1 {
				pair.buyStack = pair.buyStack[1:]
			} else {
				pair.buyStack = pair.buyStack[:0]
			}

		} else if buyerWant <= orderToSell.Curr1volume &&
			sellerWant <= orderToBuy.Curr2volume {
			fmt.Println("Case 3")

			if orderToBuy.IsMarketPrice {
				orderToSell.Curr2volume += sellerWant
				orderToBuy.Curr2volume -= sellerWant

				orderToBuy.Curr1volume += orderToSell.Curr1volume
				orderToSell.Curr1volume = 0
			} else {
				orderToSell.Curr2volume += sellerWant
				orderToSell.StockBonus += (orderToBuy.Curr2volume - sellerWant)
				orderToBuy.Curr2volume = 0

				orderToBuy.Curr1volume += buyerWant
				orderToBuy.StockBonus += (orderToSell.Curr1volume - buyerWant)
				orderToSell.Curr1volume = 0
			}
			orderToSell.IsClose = true
			if orderToSell.IsMarketPrice {
				orderToSell.Curr2volume += orderToSell.StockBonus
				orderToSell.StockBonus = 0
			}

			if len(pair.sellStack) > 1 {
				pair.sellStack = pair.sellStack[1:]
			} else {
				pair.sellStack = pair.sellStack[:0]
			}

			if orderToBuy.IsMarketPrice {
				orderToBuy.Curr1volume += orderToBuy.StockBonus
				orderToBuy.StockBonus = 0
			}
			if orderToBuy.Curr2volume == 0 {
				orderToBuy.IsClose = true

				if len(pair.buyStack) > 1 {
					pair.buyStack = pair.buyStack[1:]
				} else {
					pair.buyStack = pair.buyStack[:0]
				}
			}

		}

		m.lastEventId++
		ev1 := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Change,
			Order:     &orderToBuy,
		}
		m.lastEvents = append(m.lastEvents, ev1)
		m.lastEventId++
		ev2 := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Change,
			Order:     &orderToSell,
		}
		m.lastEvents = append(m.lastEvents, ev2)

		m.closeOrders(pair)
	}

}

func (m *Market) closeByMarket(marketOrder *Order) {
	pair := marketOrder.pair
	if marketOrder.IsGreen {

		if marketOrder.Curr2volume == 0 {
			marketOrder.IsClose = true

		} else if len(pair.sellStack) == 0 {

			fmt.Println("Sell stack is empty")
			// TODO: create order to buy

		} else {
			orderToSell := pair.sellStack[0]

			sellerWant := orderToSell.Curr2 - orderToSell.Curr2volume

			if sellerWant < marketOrder.Curr2volume {
				marketOrder.Curr2volume -= sellerWant
				orderToSell.Curr2volume += sellerWant
				marketOrder.Curr1volume += orderToSell.Curr1volume
				orderToSell.Curr1volume = 0
			} else {

			}

			m.lastEventId++
			ev1 := Event{
				Id:        m.lastEventId,
				Time:      time.Now().UnixNano(),
				EventType: Change,
				Order:     &orderToSell,
			}
			m.lastEvents = append(m.lastEvents, ev1)

		}

		m.lastEventId++
		ev1 := Event{
			Id:        m.lastEventId,
			Time:      time.Now().UnixNano(),
			EventType: Change,
			Order:     marketOrder,
		}
		m.lastEvents = append(m.lastEvents, ev1)

		//} else {
		//	//orderToSell := order
		//	fmt.Println(orderToSell)
		//}
	}
}

// TODO: правильно считать цену закрытия. Это важно
// TODO: отдельная функция при покупке по рыночной цене
// TODO: сигнал о досрочном закрытии ордера
// TODO: специальные переменные для lastprice и общего обьема валюты в паре
// TODO: корректная работа рынка с пустыми стаканами
