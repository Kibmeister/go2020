package main

import ( "net" ; "fmt" ; "bufio" ; "strings" ; "os" ; "encoding/gob" ; "io" ; "sync" ; "sort" ; "strconv" ; "time" )

type Message struct {
	Connections []string
	Transaction *Transaction
	Broadcast string
}

type Ledger struct {
	Accounts map[string]int
	Connections [] string
	Transactions[] *Transaction

}

type Transaction struct {
	Id string
	Amount int
	From string
	To string
}

var connections []net.Conn

var connectedPeers []string

var myIp string

var localMsg *Message

var ledger *Ledger

var mutex sync.Mutex

var isBroadcasted map[string]bool

var transactionIsUsed map[string]bool

func makeMessage() *Message {
	Message := new(Message)
	return Message
}

func makeLedger() *Ledger {
	Ledger := new(Ledger)
	Ledger.Accounts = make(map[string]int)
	return Ledger
}

func makeTransaction() *Transaction {
	Transaction := new(Transaction)
	return Transaction
}

func send(msg *Message, conn net.Conn) {
	enc := gob.NewEncoder(conn) // encodes connection
	enc.Encode(msg) // encodes and sends array
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// struct for recieving array
	message := &Message{}

  for {
		dec := gob.NewDecoder(conn) //decodes on the connection
		err := dec.Decode(message) // decodes the stringarray

		// checks if Connections list is not emty - if not it will be treated as a new peer
		if message.Connections != nil {
			checkForConnections(message) // check if the peer is connected
			connectToNewPeers() // connect the peers <10
			broadcast(myIp) // broadcast presence
		}

		// checks if transactions list is not emty
		if message.Transaction != nil {
			incommingTransaction(message.Transaction)
		}
				
		if (err == io.EOF) {
			fmt.Println("Connection has been closed by " + conn.RemoteAddr().String())
			return
		}

		// checks if the connections has been broadcasted
		if message.Broadcast != "" {
			mutex.Lock()
			_, exists := isBroadcasted[message.Broadcast]
			if !exists {
				isBroadcasted[message.Broadcast] = false
				fmt.Println("Is present: " + message.Broadcast)
				isBroadcasted[message.Broadcast] = true
				broadcast(myIp)
			} 
			mutex.Unlock()
		}
	} 
}

// handles outgoing transactions
func handleTransaction(t *Transaction) {
	mutex.Lock()
	exists := transactionIsUsed[t.Id]
	if exists {
		return
	} else {
		localMsg.Transaction = t
		sendTransaction(t)
	}
	mutex.Unlock()
}

// handles incomming transactions
func incommingTransaction(t *Transaction) {
	mutex.Lock()
	exists := transactionIsUsed[t.Id]
	if exists {
		return
	} else {
		go sendTransaction(t)
		go executeTransaction(t)
		transactionIsUsed[t.Id] = true
	}
	mutex.Unlock()
}

func executeTransaction(t *Transaction) {
	fmt.Println("[ ID:", t.Id, ": a transaction has been recieved from ", t.From, "to ", t.To, "with the amount of", t.Amount, "]")

	// checks if there is an account with t.From
	if worth1, exists1 := ledger.Accounts[t.From]; !exists1 {
		ledger.Accounts[t.From] = 0 - t.Amount // makes an account with the transaction
	} else {
		ledger.Accounts[t.From] = worth1 - t.Amount
	}

	// checks if there is an account with t.To
	if worth2, exists2 := ledger.Accounts[t.To]; exists2 {
		ledger.Accounts[t.To] = worth2 + t.Amount
	} else {
		ledger.Accounts[t.To] = 0 + t.Amount // makes an account with the transaction
	}
	fmt.Println("worth of ", t.From, " = ", ledger.Accounts[t.From])
	fmt.Println("worth of ", t.To, " = ", ledger.Accounts[t.To])
}

// Sends a transactions to all peers
func sendTransaction(t *Transaction) {
	transaction := &Message{}
	transaction.Transaction = t
	for _, conn := range connections{
		go send(transaction, conn)
	}
	transactionIsUsed[t.Id] = true
}

// sorts and checks peers
func checkForConnections(msg *Message) {
	mutex.Lock()
	sort.Strings(msg.Connections)
	for _, peer := range msg.Connections{
		exists, _ := findString(localMsg.Connections, peer)
		if !exists {
			localMsg.Connections = append(localMsg.Connections, peer)
		}
	}
	mutex.Unlock()
}

// connect to newcommers
func connectToNewPeers() {
	mutex.Lock()
	if len(connectedPeers)  > 10 { //if legnth of peers are greater than ten it will not connect to mere peers
		return
	}
	if len(connectedPeers)  < 10 {
		for _, peer := range localMsg.Connections{
			connected, _ := findString(connectedPeers, peer)
			if !connected {
				conn, _ := net.Dial("tcp", peer)
				connections = append(connections, conn)
				connectedPeers = append(connectedPeers, conn.RemoteAddr().String())
			}
		}
	}
	mutex.Unlock()
}

// help function for comparing a string to an array
func findString(list []string, p string) (bool, int){
	for ind, l := range list{
		if l == p {
			return true, ind
		}
	}
	return false, -1
}

// broadcast ip to all peers 
func broadcast(msg string) {
	isBroadcasted[msg] = false
	broadcastMsg := &Message{}
	broadcastMsg.Broadcast = msg
	for _, conn := range connections{
		go send(broadcastMsg, conn)
	}
}

// dial up peer
func dial(addr string) {
	conn, err := net.Dial("tcp", addr)
	// Checks if there is an error when dialing the conncection and connecting if possible
	if err != nil {
		fmt.Println("The peer has not been found")
		fmt.Println("Using...", myIp, "...Instead")
	} else {
		connections = append(connections, conn)
		connectedPeers = append(connectedPeers, conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}

// listening for-loop
func listeningForConnections(ln net.Listener) {
	defer ln.Close()
	for {
		incomming, _ := ln.Accept()
		connections = append(connections, incomming)
		fmt.Println("Got a connection... with" + incomming.RemoteAddr().String())
		fmt.Println("Wait for transactions...")
		go send(localMsg, incomming)
		go handleConnection(incomming)
	}
}

// requests the peer for a transaction
func requestTransaction() {
	time.Sleep(20000 *time.Millisecond) // waits 20 seconds and then requests for a transaction
	fmt.Println("Please enter a transaction with the following: from,to,amount ")
	fmt.Print(">")

	reader := bufio.NewReader(os.Stdin)
	trans, _ := reader.ReadString('\n')

	split := strings.Split(trans, ",")

	from := split[0]
	to := split[1]
	amount, err := strconv.Atoi(strings.Trim(split[2], "\r \n")) // string to int
	if err != nil {
		fmt.Println("try again")
	} else {
		t := makeTransaction()
		t.From = from
		t.To = to
		t.Amount = amount
		t.Id = myIp + from + to

		fmt.Println("[From: ", t.From, "To: ", t.To, "amount: ", t.Amount, "id: ", t.Id)
		go handleTransaction(t)
	}
}

func main() {
	// initiating local variables
	localMsg = makeMessage()

	ledger = makeLedger()

	// stores the broadcasts and is true if used
	isBroadcasted = make(map[string]bool)

	// stores the transactions and is true if used
	transactionIsUsed = make(map[string]bool)

	// starting reader on terminal
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write IP and port")
	fmt.Print(">")

	// address which to dial
	peerAddr, _ := reader.ReadString('\n')
	peerAddr = strings.TrimSpace(peerAddr) // returns peerAdrr without trailing suffix

	// Opening a listener with a random port
	ln, _ := net.Listen("tcp", ":" )
	myIp = ln.Addr().String()
	localMsg.Connections = append(localMsg.Connections, myIp)
	connectedPeers = append(connectedPeers, myIp)
	fmt.Println("Listening on...", myIp)

	// accepter for new connections
	go listeningForConnections(ln)

	// dialing up the peeradress
	go dial(peerAddr)

	// foreloop that keeps the system going
	for {
		requestTransaction()
	}
}
