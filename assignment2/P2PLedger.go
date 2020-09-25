package main

import ( "net" ; "fmt" ; "bufio" ; "strings" ; "os" ; "sync" ; "encoding/gob" ; "io" ; "log")

// TODO: remove the sting and brodcast functionlity 
// TODO: 
// TODO: implement a Ledger
// TODO: implement a Tansaction functionallity so clients can make transactions
// TODO: system should ensure eventual consistency 
// TODO: Work in two settings : all the peers connect and after a long break make tansactions
// TODO: 											: for late comers, make a list of all the transactions and forward them
// TODO: 											: to clients that join later. 


var peers []net.Conn

var mutex sync.Mutex

var messagesSent =  make(map[*Ledger]bool)

var haveSent = make(map[string]string)

var requester = make(map[string]bool)

type Ledger struct {
	Accounts map[string]int
	MessagesSent string
	Connections []string
}

type Transaction struct {
	Id string
	Amount int
	From string
	To string
}

func makeTransaction() *Transaction {
	Transaction := new(Transaction)
	return Transaction
}

func makeLedger() *Ledger {
	Ledger := new(Ledger)
	return Ledger
}

func dummy() {
	Ledger := makeLedger()
	var t1 = makeTransaction()
	t1.Id = "Lars"
	t1.Amount = 12
	t1.From = "Lars"
	t1.To = "Birgit"
	Ledger.Accounts[t1.From] = 100


	var t2 = makeTransaction()
	t2.Id = "Birgit"
	t2.Amount = -13
	t2.From = "Birgit"
	t2.To = "Lars"
	Ledger.Accounts[t2.From] = 1234

	var t3 = makeTransaction()
	t3.Id = "Lise"
	t3.Amount = 155
	t3.From = "Lise"
	t3.To = "Lars"
	Ledger.Accounts[t1.From] = 1245


	for i := 0; i < 100; i++ {
		// call transaction method
	}
}

func (l *Ledger) Transaction(t *Transaction) {
	mutex.Lock()
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
	mutex.Unlock()
}

// update the connectionsList
func UpdateLedger() *Ledger {
	mutex.Lock()
	Ledger := new(Ledger)
	for _, peer := range peers {
		Ledger.Connections = append(Ledger.Connections, peer.RemoteAddr().String())
	}
	Ledger.Connections = append(Ledger.Connections, peers[0].LocalAddr().String())
	Ledger.MessagesSent = "peerlist"
	mutex.Unlock()
	return Ledger
}

// Send the ledger to the peers connected
func sendLedger(conn net.Conn) {
	ledger := UpdateLedger() // string array
	enc := gob.NewEncoder(conn) // encodes connection
	enc.Encode(ledger) // encodes and sends array
}

// Dial other connections
func dialNewConnections(peersToConnect[] string){
	mutex.Lock()
	for _, peerC := range peersToConnect {
		conn, err := net.Dial("tcp", peerC)
		if err != nil {
		} else {
			handleConnection(conn)
		}
	}
	mutex.Unlock()
}

func handleMessage(ledger *Ledger, conn net.Conn) {
	var u string = ledger.MessagesSent
	mutex.Lock()
	if u == "RequestingPeerList"{
		requester[conn.RemoteAddr().String()] = true
		go sendLedger(conn)
		fmt.Println("Have sent peerlist to" + conn.RemoteAddr().String())
		mutex.Unlock()
		return
	} 
	if u == "peerlist" && requester[conn.RemoteAddr().String()] == true{
		fmt.Println("got peerlist from")
		fmt.Println(ledger.Connections)
		go dialNewConnections(ledger.Connections)
		mutex.Unlock()
		return
	} else{
		mutex.Unlock()
		return
	} 
}

func Requester(conn net.Conn, m string) {
	mutex.Lock()
	ledger := new(Ledger)
	ledger.MessagesSent = m
	enc := gob.NewEncoder(conn) // encodes connection
	enc.Encode(ledger) // encodes and sends array
	mutex.Unlock()
}


func handleConnection(conn net.Conn) {
	defer conn.Close()

	haveSent := make(map[string]string)
	// adding conn to the networklist
	peers = append(peers, conn)

	//request for connections peer list
	go Requester(conn, "RequestingPeerList")

	// struct for recieving array
	recieved := &Ledger{}

  for {
			dec := gob.NewDecoder(conn) //decodes on the connection
			err := dec.Decode(recieved) // decodes the stringarray

			if haveSent[recieved.MessagesSent] != conn.RemoteAddr().String() {
				haveSent[recieved.MessagesSent] = conn.RemoteAddr().String()
				fmt.Println("wat1")
				fmt.Println(recieved.MessagesSent)
				go handleMessage(recieved, conn) // Handling the queries
				
				if (err == io.EOF) {
					fmt.Println("Connection closed by " + conn.RemoteAddr().String())
					return
				}

				if (err != nil) {
					log.Println(err.Error())
					return
				}
			} else {
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
	ln, _ := net.Listen("tcp", ":" )
	_, port, _ := net.SplitHostPort(ln.Addr().String()) //splits the port "host:port"


	//Listing and accepting incoming clients
	fmt.Println("Listening on port: " + port)
	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection... with" + conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}