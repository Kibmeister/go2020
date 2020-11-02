package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"crypto/sha256"
	//"time"
)

var e = big.NewInt(3)
var n = big.NewInt(0)        // public key
var d = big.NewInt(0)        // privat key
var encrypted = new(big.Int) // encrypted message new.Big.NewInt(0)
var decrypted = new(big.Int) // decrypted message
var signedMessage = new(big.Int)
var byteMessage []byte 


// returns hashed bytes of string in big.Int
func Hash(Message []byte) []byte {
	h := sha256.New()
	_, err := h.Write([]byte(Message))
	if err != nil {
		panic(err)
	}
	// fmt.Println("this is the hash sum : ", h.Sum(nil) )
	return h.Sum(nil)
}

//signs message by decrypting the hashed message
func SignMessage(Message []byte) *big.Int {
	msgSum := Hash(Message)
	byteMessage = msgSum

	signed := new(big.Int).Exp(new(big.Int).SetBytes(msgSum), d, n)
	signedMessage = signed
	return signed
}

// 
func verify() {
	fmt.Println("signatur : " , new(big.Int).Exp(signedMessage, e, n).Int64())
	fmt.Println("comparer : ", new(big.Int).SetBytes(byteMessage).Int64())
	if new(big.Int).SetBytes(byteMessage).Int64() == new(big.Int).Exp(signedMessage, e, n).Int64() {
		fmt.Println("verification is goood ")
	} else {
		fmt.Println("this is not verified ")
	}
}

func KeyGen(k int) {
	for {
		p, _ := rand.Prime(rand.Reader, k/2)   // makes a random prime
		q, _ := rand.Prime(rand.Reader, k-k/2) // makes a random prime
		// fmt.Println("the first random prime : ", p)
		// fmt.Println("the second random prime : ", q)

		n = n.Mul(p, q) // first digit in our public key

		e := big.NewInt(3) // the exponent to match with the public key

		one := big.NewInt(1)

		pSub := new(big.Int).Sub(p, one)
		qSub := new(big.Int).Sub(q, one)
		pq := new(big.Int).Mul(pSub, qSub)

		d.ModInverse(e, pq)
		integer := new(big.Int).Mul(e, d)
		integer.Mod(integer, pq)

		if integer.Int64() == 1 {
			fmt.Println("this is the publicKey", n)
			fmt.Println("this is the privateKey :", d)
			return
		}
	}
}

// Test your solution by verifying (at least) that 
// your modulus has the required length and that
// encryption followed by decryption of a few random
// plaintexts outputs the original plaintexts. Note that 
//plaintexts and ciphtertexts in RSA are basically numbers
// in a certain interval. So it is sufficient to test if
// encryption of a number followed by decryption returns 
// the original number. You do not need to, for instance, 
// convert character strings to numbers.

func encrypt(m *big.Int, e *big.Int, n *big.Int) *big.Int {
	fmt.Println("the number before encryption :", m)
	encrypted.Exp(m, e, n)

	fmt.Println("this is encrypted :", encrypted)

	return encrypted
}

func decrypt(c *big.Int, d *big.Int, n *big.Int) *big.Int {

	fmt.Println("encrypt : ", encrypted, "public key : ", n, "private key : ", d)
	decrypted.Exp(c, d, n)

	fmt.Println("this is decrypted :", decrypted)

	return decrypted
}

// function for generating the file where the secretkey can be encrypted under a password
func Generate(filename string, password string) string {
	KeyGen(52)
	var pk string // publickey
	var sk string // secretkey
	pk = n.String() + ", " + e.String() // gathers pk in a single string
	sk = d.String()

	slice := make([]byte, 16) // the secretkey has to have a blocklength of 16, because of iv
	copy(slice, password)
	plaintext := []byte(sk)

	block, err := aes.NewCipher(slice)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[aes.BlockSize:]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filename, ciphertext, 0777)
	if err != nil {
		panic(err)
	}
	d = big.NewInt(0)
	return pk
}

func Sign(filename string, password string, msg []byte) *big.Int {
	var sk string
	newD := new(big.Int)
	signature := new(big.Int)
	sk = decryptFromFile(password, filename) // return secret key the program does not work 
	newD = stringToBigInt(sk)
	d = newD
	signature = SignMessage(msg)
	return signature
}

func stringToBigInt(s string) *big.Int {
	i := new(big.Int)
    _, err := fmt.Sscan(s, i)
    if err != nil {
        fmt.Println("error scanning value:", err)
    } else {
				return i
		}
	return i
}

/*func encryptToFile(key string, encrypted string, file string) {
	// Input = filename and write ciphertext to file
	slice := make([]byte, 16)
	copy(slice, key)
	plaintext := []byte(encrypted)

	block, err := aes.NewCipher(slice)
	if err != nil {
		panic(err)
	}
	
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[aes.BlockSize:]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(file, ciphertext, 0777)
	if err != nil {
		panic(err)
	}
}*/

func decryptFromFile(key string, file string) string {
	// read ciphertext from file --> decrypt --> output plaintext
	
	slice := make([]byte, 16)
	copy(slice, key)

	block, err := aes.NewCipher(slice)
	if err != nil {
		fmt.Println("error occured")
		panic(err)
	}
	ciphertext, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("error occured")
		panic(err)
	}

	plaintext := make([]byte, aes.BlockSize+len(ciphertext))
	iv := plaintext[:aes.BlockSize]

	//CBC mode always works in whole blocks.
	mode := cipher.NewCTR(block, iv)
	mode.XORKeyStream(plaintext, ciphertext[aes.BlockSize:] )

	sk := string(plaintext)
	fmt.Printf("%s\n", plaintext)
	fmt.Println(string(plaintext), "this is D")
	return sk
}

func main() {
	//start := time.Now()
	//KeyGen(52)
	//sign([]byte("message"))
	//verify()
	//encrypt(big.NewInt(77), e, n)
	//decrypt(encrypted, d, n)
	//cipher := "TheSecretMessage"
	//encrypted := "file.txt"  //OPRET EN FIL I MAPPEN MED DETTE NAVN
	//privateKey := d.String() //d er værdien ved en privatekey
	//encryptToFile("hello", cipher, encrypted)
	//decryptFromFile("hello", encrypted)
	Generate("filename", "password")
	Sign("filename", "pasword", []byte("hello you"))
	//fmt.Println(time.Since(start))
}
