package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"crypto/sha256"
)

var e = big.NewInt(3)
var n = big.NewInt(0)        // public key
var d = big.NewInt(0)        // privat key
var encrypted = new(big.Int) // encrypted message new.Big.NewInt(0)
var decrypted = new(big.Int) // decrypted message
var signedMessage = new(big.Int)

func Hash(Message []byte) *big.Int {
	hash := new(big.Int)
	h := sha256.New()
	_, err := h.Write([]byte(Message))
	if err != nil {
		panic(err)
	}
	hash.SetBytes(h.Sum(nil))
	return hash
}

func sign(Message string) {
	msg := []byte(Message)
	msgSum := Hash(msg)
	signed := decrypt(msgSum, e, n)
	signedMessage = signed
}

func verify(m *big.Int,) {
	verify := encrypt(m, d, n)
	message := Hash([]byte("this is a message"))
	if verify.Cmp(message) == 0 {
		fmt.Println("message verified")
	}
	if verify.Cmp(message) != 0 {
		fmt.Println("message not verified")
	}
}

func keyGen(k int) {
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

// Test your solution by verifying (at least) that your modulus has the required length and that encryption followed by decryption of a few random plaintexts outputs the original plaintexts. Note that plaintexts and ciphtertexts in RSA are basically numbers in a certain interval. So it is sufficient to test if encryption of a number followed by decryption returns the original number. You do not need to, for instance, convert character strings to numbers.

func encrypt(m *big.Int, e *big.Int, n *big.Int) *big.Int {
	fmt.Println("the number before encryption :", m)
	encrypted.Exp(m, e, n)

	fmt.Println("this is encrypted :", encrypted)

	return encrypted
}

func decrypt(c *big.Int, d *big.Int, n *big.Int) *big.Int {

	fmt.Println("encrypt : ", encrypted, "public key : ", n, "private key : ", d)
	decrypted.Exp(encrypted, d, n)

	fmt.Println("this is decrypted :", decrypted)

	return decrypted
}

func encryptToFile(key string, encrypted string, file string) {
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
}

func decryptFromFile(key string, file string) {
	// read ciphertext from file --> decrypt --> output plaintext
	
	slice := make([]byte, 16)
	copy(slice, key)

	block, err := aes.NewCipher(slice)
	if err != nil {
		panic(err)
	}
	ciphertext, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	plaintext := make([]byte, aes.BlockSize+len(ciphertext))
	iv := plaintext[:aes.BlockSize]

	//CBC mode always works in whole blocks.
	mode := cipher.NewCTR(block, iv)
	mode.XORKeyStream(plaintext, ciphertext[aes.BlockSize:] )

	fmt.Printf("%s\n", plaintext)
}

func main() {
	keyGen(64)
	sign("this is a message")
	verify(signedMessage)
	encrypt(big.NewInt(77), e, n)
	decrypt(encrypted, d, n)
	cipher := "TheSecretMessage"
	encrypted := "file.txt"  //OPRET EN FIL I MAPPEN MED DETTE NAVN
	// privateKey := d.String() //d er værdien ved en privatekey
	encryptToFile("hello", cipher, encrypted)
	decryptFromFile("hello", encrypted)
}