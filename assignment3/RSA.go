package main 

import ( "fmt" ; "math/big" ; "crypto/rand" )

func keyGen(k int){
	p, _ := rand.Prime(rand.Reader, k/2)
	q, _ := rand.Prime(rand.Reader, k/2)

	s := new(big.Int)
	fmt.Println(p)
	fmt.Println(q)
	
	var n = s.Mul(p,q)
	fmt.Println(n)
}

func main() {
	keyGen(124)
}