package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var e = big.NewInt(3)
var n = big.NewInt(0)         // public key
var d = big.NewInt(0)         // privat key
var encrypted = big.NewInt(0) // encrypted message
var decrypted = big.NewInt(0) // decrypted message

type PublicKey struct {
	PublicKeyE *big.Int
	PublicKeyN *big.Int
}

func makePublicKey() *PublicKey {
	PublicKey := new(PublicKey)
	return PublicKey
}

func keyGen(k int) *PublicKey {
	PublicKey := makePublicKey()
	p, _ := rand.Prime(rand.Reader, k/2)   // makes a random prime
	q, _ := rand.Prime(rand.Reader, k-k/2) // makes a random prime
	fmt.Println("the first random prime : ", p)
	fmt.Println("the second random prime : ", q)

	n = n.Mul(p, q) // first digit in our public key
	PublicKey.PublicKeyN = n

	e := big.NewInt(3) // the exponent to match with the public key
	PublicKey.PublicKeyE = e

	one := big.NewInt(1)

	primesSubtract := new(big.Int)

	primesSubtract = primesSubtract.Mul(primesSubtract.Sub(p, one), primesSubtract.Sub(q, one))

	d = e.ModInverse(e, primesSubtract)

	// gcdP := new(big.Int).GCD(nil, nil, e, p.Sub(p, one)) // the greates common divisor for p
	// fmt.Println("gcd for p: ", gcdP)
	// gcdQ := new(big.Int).GCD(nil, nil, e, q.Sub(q, one)) // the greates common divisor for q
	// fmt.Println("gcd for q: ", gcdQ)

	// if gcdP == big.NewInt(1) && gcdQ == big.NewInt(1)  {
	// 	d = e.ModInverse(e, primesSubtract)
	// 	fmt.Println("this is the privateKey :", d)
	// }
	fmt.Println("this is the publicKey", PublicKey.PublicKeyN, PublicKey.PublicKeyE)
	fmt.Println("this is the privateKey :", d)

	return PublicKey
}

// Test your solution by verifying (at least) that your modulus has the required length and that encryption followed by decryption of a few random plaintexts outputs the original plaintexts. Note that plaintexts and ciphtertexts in RSA are basically numbers in a certain interval. So it is sufficient to test if encryption of a number followed by decryption returns the original number. You do not need to, for instance, convert character strings to numbers.

func encrypt(m *big.Int, n *big.Int, e *big.Int) *big.Int {
	// create a range that stores the cifers of i
	bigInt := new(big.Int)

	bigInt.Exp(m, e, n)
	encrypted = bigInt
	fmt.Println("this is encrypted :", encrypted)

	return bigInt
}

func decrypt(encrypted *big.Int, n *big.Int, d *big.Int) *big.Int {
	bigInt := new(big.Int)
	fmt.Println("encrypt : ", encrypted, "public key : ", n, "private key : ", d)
	bigInt.Exp(encrypted, d, n)
	decrypted = bigInt
	fmt.Println("this is decrypted :", decrypted)

	return bigInt
}

func main() {
	keyGen(9)
	encrypt(big.NewInt(20), n, e)
	decrypt(encrypted, n, d)
}
