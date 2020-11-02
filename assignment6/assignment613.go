package main

import ( 
	"net" 
	"fmt"  
	"bufio"  
	"strings"  
	"os"  
	"encoding/gob"  
	"io"  
	"sync"  
	"sort"  
	"strconv"  
	"time" 
	"math/big"
	RSA "./RSA"
)

type Message struct {
	Connections []string
	Transaction *Transaction
	Broadcast string
	Ledger *Ledger
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
	Signature string
}

var connections []net.Conn

var connectedPeers []string

var myIp string

var localMsg *Message

var ledger *Ledger

var mutex sync.Mutex

var isBroadcasted map[string]bool

var transactionIsUsed map[string]bool

var n,e *big.Int

var LocalPK *RSA.PublicKeyPair

var PKlist[] *RSA.PublicKeyPair

var PkString string

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

		// when message.ledger not empty, this will checked if the localledger aldready have the same pk. If not, this pk will be added to the local ledger
		if(message.Ledger != nil) {
			addNewPk(message)
		}

		// checks if transactions list is not emty
		if message.Transaction != nil {
			//fmt.Println("thisis peers:", connectedPeers)
			checkIfUserIsKnown(message, conn) // if the ip is know the PublicKey is added to the local accounts list
			incommingTransaction(message.Transaction) // handles incomming transaction
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

// checks connections of sender if it is known
func checkIfUserIsKnown(m *Message, conn net.Conn){
	exists, _ := findString(connectedPeers, conn.RemoteAddr().String())
	if exists {
		addNewPk(m)
	} else {
		return
	}
}

// add users pk to local ledger
func addNewPk(m *Message) {
	for key, element := range m.Ledger.Accounts {
		if _, exists1 := ledger.Accounts[key]; !exists1 {
			ledger.Accounts[key] = element
			sortKeyPair(key) // sorts the publicKey
			//fmt.Println("this the key: ",key)
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
		//localMsg.Transaction = t
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
		verifyTransaction(t)
		transactionIsUsed[t.Id] = true
	}
	mutex.Unlock()
}

func executeTransaction(t *Transaction) {
	fmt.Println("------------------------TRANSACTION RECIEVED----------------------------------")
	fmt.Println("[ ID:", t.Id, ": a transaction has been recieved from ", t.From, "to ", t.To, "with the amount of", t.Amount, "]")
	fmt.Println("------------------------------------------------------------------------------")

	// checks if there is an account with t.From                                  
	if worth1, exists1 := ledger.Accounts[t.To]; !exists1 {
		fmt.Println("account is not recognizable")
	} else {
		ledger.Accounts[t.From] = worth1 - t.Amount
		ledger.Accounts[t.To] += t.Amount
	}
}

// Sends a transactions to all peers
func sendTransaction(t *Transaction) {
	transactionToSend := &Message{}
	transactionToSend.Transaction = t
	for _, conn := range connections{
		go send(transactionToSend, conn)
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
			localMsg.Ledger = ledger
			addNewPk(msg)
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

// sorts the PublicKeys from other connections and them to localPK list in a ledger containing Pks.
func sortKeyPair(pk string) *RSA.PublicKeyPair {
	u := new(RSA.PublicKeyPair)
	split := strings.Split(pk, ",")
	//fmt.Println(pk)
	n2 := new(big.Int)
	e2 := new(big.Int)
	n2 = stringToBigInt(split[0])
	e2 = stringToBigInt(split[1])
	//fmt.Println("n1:", n2)
	//fmt.Println("e1:", e2)
	u.N = n2
	u.E = e2
	PKlist = append(PKlist, u)
	//fmt.Println(PKlist)
	return u
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
		// addPk to localMessage
		localMsg.Ledger = ledger
		fmt.Println("Got a connection... with" + incomming.RemoteAddr().String())
		fmt.Println("Wait for transactions...")
		go send(localMsg, incomming)
		go handleConnection(incomming)
	}
}

// requests the peer for a transaction
func requestTransaction() {
	time.Sleep(20000 *time.Millisecond) // waits 20 seconds and then requests for a transaction

	
	// promt receivers with their following public kyes 
	fmt.Println("Please enter a transaction by following order: to,amount.")
	fmt.Println("------------------------------------------------------")
	fmt.Println("The disired reciever, to, can be chosen by typing the index on the left:")
	fmt.Println("------------------------------------------------------")
	

	var localMap = make(map[int]string)
	var i = 0

	for key, _ := range ledger.Accounts {
		localMap[i]	= key
		fmt.Println(i,": ", key)
		fmt.Println("------------------------------------------------------")
		i++
	}

	fmt.Println("->")
	
	var toSign string

	reader := bufio.NewReader(os.Stdin)
	trans, _ := reader.ReadString('\n')

	split := strings.Split(trans, ",")

	
	to, err := strconv.Atoi(strings.Trim(split[0], "\r \n"))
	amount, err := strconv.Atoi(strings.Trim(split[1], "\r \n")) // string to int
	if err != nil {
		fmt.Println("try again")
	} else {
		t := makeTransaction()
		t.From = PkString
		t.To = localMap[to]
		t.Amount = amount
		t.Id = myIp + t.From + t.To

		toSign = t.From + t.To + strconv.Itoa(amount)

		bigIntSign := RSA.Sign([]byte(toSign))

		bigSignString := convertBigIntToString(bigIntSign)

		t.Signature = bigSignString
		fmt.Println("----------------- TRANSACTION MADE ---------------")
		fmt.Println("[From: ", t.From, "To: ", t.To, "amount: ", t.Amount, "id: ", t.Id)
		fmt.Println("--------------------------------------------------")
		go handleTransaction(t)
	}
}

// executes transaction if the results come back true
func verifyTransaction(inc *Transaction) {
	if verifySignature(inc) {
			executeTransaction(inc)
	} else {
		return
	}
}

// prepares the signature to be verified
func verifySignature(user *Transaction) bool {
	var testString string
	var strAmount string
	var testStringHash []byte
	signature := new(big.Int)
	signature = stringToBigInt(user.Signature) // recieved signature converted to big.int
	IncPK := new(RSA.PublicKeyPair) 
	strAmount = strconv.Itoa(user.Amount) 
	testString = user.From + user.To + strAmount // creates the string to get hashed
	testStringHash = []byte(testString) // created string from information
	testHash := RSA.Hash(testStringHash) // hashes the string
	//fmt.Println("this is the testHash ", testHash)
	IncPK = sortKeyPair(user.From) // sorts the keyPair to compare with signature
	if RSA.Verify(signature, testHash, IncPK) {
		return true
	} else {
		return false
	}
}

func convertBigIntToString(b *big.Int) string {
	bString := b.String()
	return bString
}

func convertPKtoString() {
	PkString = convertBigIntToString(LocalPK.N) + "," + convertBigIntToString(LocalPK.E)
}

func stringToBigInt(s string) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(s, 10)
	if !ok {
			fmt.Println("error")
			return n
	}
	return n
}

func main() {
	// initiating local variables
	localMsg = makeMessage()

	ledger = makeLedger()
	
	// generate keys for transactions
	RSA.KeyGen(1024)
	
	// initiates Public Key
	LocalPK = RSA.GetKeyPair()

	// convert pk to string
	convertPKtoString()
	
	// addPK to localledger
	ledger.Accounts[PkString] = 1000

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
