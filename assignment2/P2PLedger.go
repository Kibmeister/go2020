package main

import ( "net" ; "fmt" ; "bufio" ; "strings" ; "os" ; "sync" ; "encoding/gob" ; "io" )

// TODO: remove the sting and brodcast functionlity 
// TODO: 
// TODO: implement a Ledger
// TODO: implement a Tansaction functionallity so clients can make transactions
// TODO: system should ensure eventual consistency 
// TODO: Work in two settings : all the peers connect and after a long break make tansactions
// TODO: 											: for late comers, make a list of all the transactions and forward them
// TODO: 											: to clients that join later. 


var peers []net.Conn

var peermutex sync.Mutex

var mutex sync.Mutex

var SenderMutex sync.Mutex

var MyConnection string

var peerLedger Ledger //peermutex

var Transactions Transaction

var Accounts Ledger 

var requester = make(map[string]int)

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

func makeTransaction() *Transaction {
	Transaction := new(Transaction)
	return Transaction
}

func (l *Ledger) Transaction(t *Transaction) {
	mutex.Lock()
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
	mutex.Unlock()
}

func transactionHandler(conn net.Conn){
	t := &Transaction{}
	for {
		dec := gob.NewDecoder(conn) //decodes on the connection
		err := dec.Decode(t) // decodes the stringarray
		if (err == io.EOF) {
			fmt.Println("Connection closed by " + conn.RemoteAddr().String())
			return
		}
	}
}

func makeLedger() *Ledger {
	Ledger := new(Ledger)
	return Ledger
}

func makeMessage() *Message {
	Message := new(Message)
	return Message
}


// update the connectionsList
func UpdateConnections() *Message {
	mutex.Lock()
	peerLedger := makeMessage()
	for _, peer := range peers {
		peerLedger.Connections = append(peerLedger.Connections, peer.RemoteAddr().String())
	}
	peerLedger.Connections = append(peerLedger.Connections, MyConnection)
	fmt.Println(peerLedger.Connections)
	mutex.Unlock()
	return peerLedger
}

// Send the ledger to the peers connected
func sendPeerList(conn net.Conn) {
	c := makeMessage()
	c.Connections = peerLedger.Connections // string array
	fmt.Println(peerLedger.Connections, "peer")
	fmt.Println(c.Connections)
	enc := gob.NewEncoder(conn) // encodes connection
	enc.Encode(c) // encodes and sends array
}

// Dial other connections
func dialNewConnections(peersToConnect[] string){
	peermutex.Lock()
	fmt.Println("call")
	for _, peerC := range peersToConnect {
		check, _ := findString(peerLedger.Connections, peerC)
		if check {
			fmt.Println("peer already connected")
		}
		if !check {
			conn, err := net.Dial("tcp", peerC)
			if err != nil {
				fmt.Println("The peer has not been found")
			} 
			if err == nil{
				go handleConnection(conn)
			}
		}
	}
	peermutex.Unlock()
}
	

func findString(list []string, p string) (bool, int){
	for ind, l := range list{
		if l == p {
			return true, ind
		}
	}
	return false, -1
}

func handleMessage(m *Message) {
	fmt.Println(2)
	mutex.Lock()
	if m.Connections != nil {
		fmt.Println(3)
		fmt.Println(m.Connections)
		//go dialNewConnections(m.Connections)
		mutex.Unlock()
	} else {
		mutex.Unlock()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	//request for connections peer list
	fmt.Println("1")

	// struct for recieving array
	msg := &Message{}

  for {
		dec := gob.NewDecoder(conn) //decodes on the connection
		err := dec.Decode(msg) // decodes the stringarray
		go handleMessage(msg) // Handling the queries
				
		if (err == io.EOF) {
			fmt.Println("Connection closed by " + conn.RemoteAddr().String())
			return
		}
	} 
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write IP and port")
	fmt.Print(">")

	// address which to dial
	peerAddr, _ := reader.ReadString('\n')
	peerAddr = strings.TrimSpace(peerAddr) // returns peerAdrr without trailing suffix

	// Calling the peerAddr
	conn, err := net.Dial("tcp", peerAddr)

	// Checks if there is an error when dialing the conncection
	if err != nil {
		fmt.Println("The peer has not been found")
		fmt.Println("Creating network...")
	} else {
		go handleConnection(conn)
	}

	// Opening a port with a random port
	ln, err := net.Listen("tcp", ":" )

	if err != nil {
		fmt.Println(err)
	}
	defer ln.Close()
	_, port, _ := net.SplitHostPort(ln.Addr().String()) //splits the port "host:port"

	MyConnection = ln.Addr().String()

	fmt.Println(MyConnection)
	

	//Listing and accepting incoming clients
	fmt.Println("Listening on port: " + port)

	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection... with" + conn.RemoteAddr().String())
		peers = append(peers, conn)
		go UpdateConnections()
		go handleConnection(conn)
		go sendPeerList(conn)
	}
}