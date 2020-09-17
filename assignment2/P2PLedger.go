package main

import ( "net" ; "fmt" ; "bufio" ; "strings" ; "os" ; "sync" )

var peers []net.Conn
var mutex sync.Mutex

var messagesSent =  make(map[string]bool)
var c = make(chan string)

// TODO: remove the sting and brodcast functionlity 
// TODO: 
// TODO: implement a Ledger
// TODO: implement a Tansaction functionallity so clients can make transactions
// TODO: system should ensure eventual consistency 
// TODO: Work in two settings : all the peers connect and after a long break make tansactions
// TODO: 											: for late comers, make a list of all the transactions and forward them
// TODO: 											: to clients that join later. 


func handleConnection(conn net.Conn) { 
	defer conn.Close()

	reader := bufio.NewReader(conn)

	otherEnd := conn.RemoteAddr().String()

  for {
		msg, err := reader.ReadString('\n')
		if (err != nil) {
			fmt.Println("Ending session with " + otherEnd)
			return
		}

		fmt.Print(string(msg))
		c <- msg
	}
}

func userInput(){
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _ := reader.ReadString('\n')
		c <- msg
	}
}

func writeMessage(){
	msg := <- c
	for _, peer := range peers {
		if !messagesSent[msg] {
			peer.Write([]byte(msg))
			messagesSent[msg] = true
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

	go userInput()
	
	go writeMessage()

	// Checks if there is an error when dialing the conncection
	if err != nil {
		fmt.Println("The peer has not been found")
		fmt.Println("Creating network...")
	} else {
		peers = append(peers, conn)
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
		peers = append(peers, conn)
		go handleConnection(conn)
	}
}