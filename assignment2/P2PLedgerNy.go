package main

type Message struct {
	Connections []string
	Transaction *Transaction
	Broadcast string
}

type Ledger struct {
	Accounts map[string]int
	Connections []string
	Transactions *Transaction

}

type Transaction struct {
	Id int
	Amount int
	From string
	To string
}

func makeMessage() *Message {
	Message := new(Message)
	return Message
}


func main() {

}