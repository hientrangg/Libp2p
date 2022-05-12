package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	glog "github.com/golang/glog"
)

func main() {
	f, err := os.Open(path)
	if err != nil {
		glog.Fatal(err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		glog.Fatal(err)
	}
	value := hex.EncodeToString(hasher.Sum(nil))
	fmt.Println(value)
}
