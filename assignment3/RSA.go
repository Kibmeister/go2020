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
	p, _ := rand.Prime(rand.Reader, k/2) // makes a random prime 
	q, _ := rand.Prime(rand.Reader, k - k/2)  // makes a random prime 
	fmt.Println("the first random prime : ", p)
	fmt.Println("the second random prime : ", q)

	n := new(big.Int) 
	
	n = n.Mul(p,q) // first digit in our public key  
	PublicKey.PublicKeyN = n
	
	e := big.NewInt(3)// the exponent to match with the public key 
	PublicKey.PublicKeyE = e

	one := big.NewInt(1)

	primesSubtract := new(big.Int)

	primesSubtract = primesSubtract.Mul(primesSubtract.Sub(p, one), primesSubtract.Sub(q, one))

	gcdP := new(big.Int).GCD(nil, nil, e, p.Sub(p, one)) // the greates common divisor for p
	gcdQ := new(big.Int).GCD(nil, nil, e, q.Sub(q, one)) // the greates common divisor for q

	if gcdP == big.NewInt(1) && gcdQ == big.NewInt(1)  {
		d := e.ModInverse(e, primesSubtract)
		fmt.Println("this is the privateKey :", d)
	}

	//fmt.Println("privatekey", d)
	fmt.Println("this is the publicKey", PublicKey.PublicKeyN, PublicKey.PublicKeyE)

	return PublicKey
}

func main() {
	keyGen(106)
}



// type keys struct {
// 	privatK *big.Int 
// 	publicK *big.Int 
// }

// func keyGen(k int){
// 	p, _ := rand.Prime(rand.Reader, k/2)
// 	q, _ := rand.Prime(rand.Reader, k - k/2)

// 	s:= new(big.Int)
// 	one:= big.NewInt(1)
// 	//fmt.Println("this is the input k :", k)

// 	e := big.NewInt(3)
// 	// generate a modulus og length |k|

// 	// this is the public key
// 	n:= s.Mul(p,q)
 

// 	sum := s.Mul(s.Sub(p, one), s.Sub(q, one))

// 	gcdP := new(big.Int).GCD(nil, nil, e, p.Sub(p, one)) // the greates common divisor for P
// 	gcdQ := new(big.Int).GCD(nil, nil, e, q.Sub(q, one)) // the greates common divisor for Q

// 	if gcdP == big.NewInt(1) && gcdQ == big.NewInt(1)  {
// 		fmt.Println("the conditions for the private key is fulfillied")
// 		d := e.ModInverse(e, sum)
// 		fmt.Println("this is my privateKey :", d)
// 	}


// 	// this is the private key
		
// 	fmt.Println("this is my publicKey  :", n)
// }

// func main() {
// 	keyGen(50)
// }


