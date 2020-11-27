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
	PK string
	Block *Block
	Genesis *GenesisBlock
}

type Ledger struct {
	Accounts map[string]int
	Connections [] string
	Transactions[] *Transaction
	PK string
	Phase int
	NewBlock *Block
}

type Block struct {
	BlockNr int
	IDlist map[int]string
	Transactions []*Transaction
	Signature string
	Creator string
}

type Transaction struct {
	Id string
	Amount int
	From string
	To string
	Signature string
}

type GenesisBlock struct {
	Accounts map[string]int
	Seed string
	Hardness *big.Int
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

var initialSeed string

var LocalBlockNumber int = 0

var tempOrder int = 0

var BlocksRecieved map[int]*Block

var blockUpdated bool = false

var isDesignatedSequencer bool = false

var SequencerKeyPair *RSA.PublicKeyPair

var informed bool = false

var slotTime int

var secretKeyList map[string]*big.Int

var localSK *big.Int

var Genesis *GenesisBlock

var BlockToBeSent[] *Block
 
var start = time.Now()

var GenesisRecieved = false

var ready = false

var blockExecuted map[*Block]bool

var printBallance = 0

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


func initiateGenesisBlock() {

	Genesis = new(GenesisBlock)
	Genesis.Accounts = make(map[string]int)
	Genesis.Seed = "LOTTERY"
	fmt.Println(Genesis.Seed)

	Genesis.Hardness = calculateHardness()

	Genesis.Accounts[PkString] = 1000000
}

func updateBlock(account string) {
	fmt.Println("Block is now updated")
	//fmt.Println(keyPair)
	var update = makeMessage()
	var sss = strconv.Itoa(LocalBlockNumber)
	var s = RSA.Sign([]byte(sss))
	var ss = convertBigIntToString(s)
	ledger.NewBlock.BlockNr = LocalBlockNumber
	ledger.NewBlock.Signature = ss
	ledger.NewBlock.Creator = account
	update.Block = ledger.NewBlock
	update.Block.Signature = ss
	//fmt.Println("This is the ledger: ", update.Ledger.NewBlock.IDlist)
	//fmt.Println(ledger.NewBlock.Signature, "this is s")
	//fmt.Println(update.Ledger.NewBlock.BlockNr, "this is the block to be send")
	//BlockToBeSent = append(BlockToBeSent, update.Ledger.NewBlock)
	
	for _, conn := range connections{
		fmt.Println("send to peers")
		go send(update, conn)
	}
	blockExecuted[update.Block] = true
	updateLocalLedger(update.Block)
	time.Sleep(1000 *time.Millisecond)
	ledger.NewBlock.IDlist = make(map[int]string)
	tempOrder = 0
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// struct for recieving array
	message := &Message{}

  for {
		dec := gob.NewDecoder(conn) //decodes on the connection
		err := dec.Decode(message) // decodes the stringarray


		// checks if Connections list is not emty - if not it will be treated as a new peer
		if ( message.Connections != nil && ledger.Phase == 1) {
			checkForConnections(message) // check if the peer is connected
			connectToNewPeers() // connect the peers <10
			broadcast(myIp) // broadcast presence
		}

		if message.PK != "" && !GenesisRecieved {
			Genesis.Accounts[message.PK] = 1000000
			fmt.Println(Genesis.Accounts)
			if len(Genesis.Accounts) == 3 {
				sendOutGenesisBlock()
				fmt.Println("sending out Genisis")
				GenesisRecieved = true
				ready = true
			}
		}

		if message.Genesis != nil && !GenesisRecieved {
			Genesis = message.Genesis
			fmt.Println("GenesisRecieved")
			GenesisRecieved = true
			ready = true
		}

		// when message.ledger not empty, this will checked if the localledger aldready have the same pk. If not, this pk will be added to the local ledger
		if message.Block != nil && ready {
			//addNewPk(message)
			fmt.Println("localBlock:", LocalBlockNumber)
			//fmt.Println("Block:", message.Ledger.NewBlock.BlockNr )
			//fmt.Println("Signature", message.Ledger.NewBlock.Signature )
			if !blockExecuted[message.Block] {
				for _, conn := range connections{
					go send(message, conn)
				}
				blockExecuted[message.Block] = true
			}

			if ( BlocksRecieved[message.Block.BlockNr] == nil && message.Block.Signature != "" ) { // if the signature is not empty, this means that there a signed transactions
				fmt.Println("this far")
				updateLocalLedger(message.Block)
			} 
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

func sendOutGenesisBlock(){
	msg := makeMessage()
	msg.Genesis = Genesis
	for _, conn := range connections{
		go send(msg, conn)
	}
}

func updateLocalLedger(b *Block) {
	mutex.Lock()
	fmt.Println("UPDATE LEDGER")
	fmt.Println("Localblock", LocalBlockNumber)
	fmt.Println("Block", b.BlockNr)
	//fmt.Println("signature", b.Signature)
	BlocksRecieved[b.BlockNr] = b // if the block can be verified, it will execute alle the transactions
	for i := 0; i<=b.BlockNr; i++ {
		if BlocksRecieved[i] != nil {
			if verifyBlock(BlocksRecieved[i]) {
				if blockExecuted[BlocksRecieved[i]] {
					Genesis.Accounts[b.Creator] = Genesis.Accounts[b.Creator] + 10 + len(b.Transactions)
					fmt.Println("The creater is now worth", Genesis.Accounts[b.Creator])
					executeAllNewTransactions(BlocksRecieved[i])
					blockExecuted[BlocksRecieved[i]] = false
				}
			}
		}
	}
	mutex.Unlock()
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

// calculates the lottery - THIS NEEDS TO BE REDACTED
func calculateLottery(account string) bool {
	slot := strconv.Itoa(slotTime) // converts slot number to string
	Hardness := Genesis.Hardness // gets hardness value
	var VK = Genesis.Seed + slot // value to sign
	var signature = RSA.Sign([]byte(VK)) //DRAW
	var draw = convertBigIntToString(signature) // converts draw to string
	var byteToHash = []byte(draw) 
	var slotToCompare = RSA.Hash(byteToHash) //hashed value of H(Seed, Slot, Vk_i, DRAW_slot,i)
	var cmd = new(big.Int)
	var hashedSum = cmd.SetBytes(slotToCompare) //Converts hashed sum to big int
	isGreaterThan := hashedSum.Cmp(Hardness) //Compares hashedSum to Hardness

	//fmt.Println(hashedSum)
	//fmt.Println(Hardness)

	if isGreaterThan == 1 { // if isGreaterThan  returns 1 - this means that the sum is greater than the hardness
		return true
	} 
	return false
}

func calculateHardness() *big.Int {
	Hardness := new(big.Int)
	a := new(big.Int).Exp(big.NewInt(2), big.NewInt(256),nil)
	b := new(big.Int).Mul(big.NewInt(94),a)
	Hardness = new(big.Int).Div(b, big.NewInt(100))
	return Hardness
}

func initateLottery() {
	time.Sleep(1000 *time.Millisecond)
	if Genesis.Accounts[PkString] > 0 {
		if calculateLottery(PkString) {
			fmt.Println("you have won")
			fmt.Println(time.Since(start))
			updateBlock(PkString)
		} else {
			time.Sleep(1000 *time.Millisecond)
			fmt.Println("not this time!")
		}
	}
	
	LocalBlockNumber = slotTime
	slotTime++
	printBallance++
	if printBallance == 10 {
		printBalance()
		printBallance = 0
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
		mutex.Unlock()
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
		mutex.Unlock()
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
		ledger.Accounts[t.To] += t.Amount-1
	}
}

// executes transactions from block
func executeAllNewTransactions(m *Block ) {
	fmt.Println("blocks has been recieved")
	tempIDList := make([]string,100)
	for i := 0; i<len(m.Transactions); i++ {
		tempIDList = append(tempIDList, m.IDlist[i])
	}
	for _, key := range ledger.Transactions {
		exists, _ := findString(tempIDList, key.Id)
		if( exists ){
			executeTransaction(key)
		}
	}
}

func createPublicKeysForTesting() {
	for i := 0; i<=2; i++ {
		RSA.KeyGen(1024)
		LocalPK = RSA.GetKeyPair()
		PkString = convertPKtoString(LocalPK)
		LocalPK = RSA.GetKeyPair()
		Genesis.Accounts[PkString] = 1000000
	}
}

// Sends a transactions to all peers
func sendTransaction(t *Transaction) {
	fmt.Println("sending transaction")
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
	if len(connectedPeers)  > 3 { //if legnth of peers are greater than ten it will not connect to mere peers
		mutex.Unlock()
		return
	}
	if len(connectedPeers)  < 3 {
		for _, peer := range localMsg.Connections{
			connected, _ := findString(connectedPeers, peer)
			if !connected {
				conn, _ := net.Dial("tcp", peer)
				connections = append(connections, conn)
				connectedPeers = append(connectedPeers, conn.RemoteAddr().String())
			}
		}
		mutex.Unlock()
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
		//initiate the genesis block
	} else {
		connections = append(connections, conn)
		connectedPeers = append(connectedPeers, conn.RemoteAddr().String())
		send(localMsg, conn)
		go handleConnection(conn)
	}
}

func initateBlock() {
	ledger.NewBlock = new(Block)
	ledger.NewBlock.IDlist = make(map[int]string)
	ledger.NewBlock.BlockNr = LocalBlockNumber
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
//51123
func getMyIp() string {
	return myIp
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
		if (t.Amount < 1) {
			fmt.Println("amount has to be at least 1 AU")
		} else {
			go handleTransaction(t)
		}
	}
}

// executes transaction if the results come back true
func verifyTransaction(inc *Transaction) {
	if verifySignature(inc) {
			ledger.NewBlock.Transactions = append(ledger.NewBlock.Transactions, inc)
			ledger.Transactions = append(ledger.Transactions, inc)
			//executeTransaction(inc)
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

func verifyBlock( b *Block ) bool {
	fmt.Println("this far222")
	var sig = b.Signature
	var sigInt = stringToBigInt(sig)
	var s = strconv.Itoa(b.BlockNr)
	//fmt.Println(b.BlockNr, "this is blocknr")
	var KP = sortKeyPair(b.Creator)
	testHash := RSA.Hash([]byte(s)) 
	//fmt.Println(KP, "this is creator")
	for _, key := range b.Transactions {
		var goesBelowZero = Genesis.Accounts[key.From] - key.Amount >= 0
		if !goesBelowZero {
			return false
		}
	}
	
	if (RSA.Verify(sigInt, testHash, KP )) {
		return true
	} else {
		return false
	}
}

func convertBigIntToString(b *big.Int) string {
	bString := b.String()
	return bString
}

func convertPKtoString(pk *RSA.PublicKeyPair) string{
	var newPkString = convertBigIntToString(pk.N) + "," + convertBigIntToString(pk.E)
	return newPkString
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
	PkString = convertPKtoString(LocalPK)

	localMsg.PK = PkString
	
	// addPK to localledger
	//ledger.Accounts[PkString] = 1000

	// stores the broadcasts and is true if used
	isBroadcasted = make(map[string]bool)

	// stores the transactions and is true if used
	transactionIsUsed = make(map[string]bool)

	// blocks
	BlocksRecieved = make(map[int]*Block)
	blockExecuted = make(map[*Block]bool)

	// secret key list
	secretKeyList = make(map[string]*big.Int)

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

	// accepter for new connections 51289
	go listeningForConnections(ln)

	// dialing up the peeradress 51230
	go dial(peerAddr)

	initiateGenesisBlock()
	//
	initateBlock()
	//initiate the genesis block

	// foreloop that keeps the system going
	for {
		if len(Genesis.Accounts) == 3 {
			//fmt.Println("sending out Genisis123123")
			initateLottery()
			//makeFakeTransaction()
		} 
		//checkForNewBlocks()
		//makeFakeTransaction()
		//requestTransaction()
		//printBallance()
	}
}

func printBalance() {
	fmt.Println("----------------Ballance------------------")
	var i = 0
	for key, _ := range Genesis.Accounts {
		fmt.Println(i, "holds the ballance of", Genesis.Accounts[key] )
		i++
	}
	fmt.Println("----------------Printed------------------")
}
 
func makeFakeTransaction() {
	t1 := makeTransaction()
	t2 := makeTransaction()

	var localMap = make(map[int]string)
	var i = 0

	for key, _ := range Genesis.Accounts {
		localMap[i]	= key
		i++
	}

	t1.From = PkString
	t1.To = localMap[0]
	t1.Amount = 11234
	t1.Id = myIp + t1.From + t1.To

	toSign1 := t1.From + t1.To + strconv.Itoa(t1.Amount)

	bigIntSign1 := RSA.Sign([]byte(toSign1))

	bigSignString1 := convertBigIntToString(bigIntSign1)

	t1.Signature = bigSignString1

	t2.From = PkString
	t2.To = localMap[1]
	t2.Amount = 129814
	t2.Id = myIp + t2.From + t2.To

	toSign2 := t2.From + t2.To + strconv.Itoa(t2.Amount)

	bigIntSign2 := RSA.Sign([]byte(toSign2))

	bigSignString2 := convertBigIntToString(bigIntSign2)

	t2.Signature = bigSignString2

	handleTransaction(t1)
	handleTransaction(t2)
}
