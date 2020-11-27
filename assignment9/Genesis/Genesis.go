package Genesis

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"crypto/sha256"
	"time"
	RSA "../RSA"
)

func initateGenesisBlock() {
	var list1 = new([]string)
	for i := 0; i<10; i++ {
		RSA.KeyGen(1024)
	}
}