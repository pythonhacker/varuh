// Test openpgp

package main

import (
	"os"
	"os/user"
	"fmt"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"golang.org/x/crypto/openpgp"
)


func main() {

	currUser, _ := user.Current()
	secretText := "These are the nuclear launch codes - A/B/C/D"
	path, err := filepath.Abs(filepath.Join(currUser.HomeDir, ".gnupg/pubring.kbx"))
	fmt.Println(path)
	
	fh, _ := os.Open(path)
	defer fh.Close()
	
	entityList, err := openpgp.ReadArmoredKeyRing(fh)
	if err != nil {
		fmt.Println("1")
		panic(err)
	}

	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, entityList, nil, nil, nil)

	_, err = w.Write([]byte(secretText))
	if err != nil {
		fmt.Println("2")		
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(buf)
    if err != nil {
		fmt.Println("3")		
		panic(err)
    }
	
	//    encStr := base64.StdEncoding.EncodeToString(bytes)
	
	err = os.WriteFile("test.gpg", data, 0644)
	if err != nil {
		fmt.Println("4")		
		panic(err)
	}
}
