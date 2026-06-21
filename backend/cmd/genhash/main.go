package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pwd := "123456"
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("HASH=%s\n", string(hash))
	if err := bcrypt.CompareHashAndPassword(hash, []byte(pwd)); err != nil {
		fmt.Println("VERIFY=FAIL", err)
	} else {
		fmt.Println("VERIFY=OK")
	}
}
