package main 

import ( "fmt" ; "math/big" ; "crypto/rand" )

var e = big.NewInt(3)

type PublicKey struct {
	PublicKeyE *big.Int
	PublicKeyN *big.Int
}

func makePublicKey() *PublicKey {
	PublicKey := new(PublicKey)
	return PublicKey
}

func keyGen(k int) *PublicKey{
	PublicKey := makePublicKey()

	p, _ := rand.Prime(rand.Reader, k/2)
	q, _ := rand.Prime(rand.Reader, k - k/2)
	fmt.Println(p)
	fmt.Println(q)

	n := new(big.Int)
	
	n = n.Mul(p,q)
	PublicKey.PublicKeyN = n
	
	
	PublicKey.PublicKeyE = e

	one := big.NewInt(1)

	sum := new(big.Int)

	sum = sum.Mul(sum.Sub(p, one), sum.Sub(q, one))

	d := new(big.Int)

	d = d.ModInverse(e, sum)

	fmt.Println("privatekey", d)
	fmt.Println("publickey", PublicKey.PublicKeyN, PublicKey.PublicKeyE)

	return PublicKey
}


func main() {
	keyGen(4)
}