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

var messagesSent =  make(map[string]bool)

type Ledger struct {
	Connections []string
}

func makeLedger() *Ledger {
	Ledger := new(Ledger)
	return Ledger
}

// update the connectionsList
func UpdateLedger() *Ledger {
	Ledger := new(Ledger)
	mutex.Lock()
	for _, peer := range peers {
		Ledger.Connections = append(Ledger.Connections, peer.RemoteAddr().String())
	}
	Ledger.Connections = append(Ledger.Connections, peers[0].LocalAddr().String())
	mutex.Unlock()
	return Ledger
}

// Send the ledger to the peers connected
func sendLedger(conn net.Conn) {
	ledger := UpdateLedger() // string array
	enc := gob.NewEncoder(conn) // encodes connection
	enc.Encode(ledger) // encodes and sends array
}

func updateLocalList(peersToConnect[] string){
	mutex.Lock()
	for _, peerC := range peersToConnect {
		fmt.Println(peerC)
		conn, err := net.Dial("tcp", peerC)
		if err != nil {
			fmt.Println("Something went wrong")
		}
		handleConnection(conn)
	}
	mutex.Unlock()
}


func handleConnection(conn net.Conn) {
	defer conn.Close()

	// adding conn to the networklist
	peers = append(peers, conn)
	go sendLedger(conn)

	// struct for recieving array
	recieved := &Ledger{}

	// reads on incomming
	//reader := bufio.NewReader(conn)

	// other end of conn
	otherEnd := conn.RemoteAddr().String()

  for {
		dec := gob.NewDecoder(conn) //decodes on the connection
		err := dec.Decode(recieved) // decodes the stringarray

		if (err == io.EOF) {
		fmt.Println("Connection closed by " + conn.RemoteAddr().String())
		return
		}

		if (err != nil) {
		log.Println(err.Error())
		return
		}

		var recieving = recieved.Connections //recieved strings of connections
		
		go updateLocalList(recieving)
		fmt.Println(recieved)
		
		if (err != nil) {
			fmt.Println("Ending session with " + otherEnd)
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
	ln, _ := net.Listen("tcp", ":" )
	_, port, _ := net.SplitHostPort(ln.Addr().String()) //splits the port "host:port"


	//Listing and accepting incoming clients
	fmt.Println("Listening on port: " + port)
	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection...")
		go handleConnection(conn)
	}
}