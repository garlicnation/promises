package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var listOfWebsites = []string{
	"https://google.com",
	"https://facebook.com",
	"https://yahoo.com",
	"https://bing.com",
	"https://github.com",
}

func main() {
	// Allocate a slice to store the contents of all of the websites we are
	// interested in.
	allContents := []byte{}
	// Record the start time of the program so that we can figure out how fast
	// it runs.
	start := time.Now()

	for _, website := range listOfWebsites {
		// Perform an HTTP GET to retrieve the website contents.
		val, err := http.Get(website)
		if err != nil {
			// We aren't interested in handling errors, so immediately exit the program if
			// an error is encountered
			panic(err)
		}
		// Next, we read the entire contents of the response's body
		contents, err := ioutil.ReadAll(val.Body)
		if err != nil {
			panic(err)
		}
		// After reading the body, we append it to our buffer before moving
		// on to the next site.
		allContents = append(allContents, contents...)
	}
	// All done, record the time again.
	end := time.Now()

	// Calculate a checksum of the responses
	checksum := sha512.Sum512(allContents)
	// Print out a human-readable form of the checksum
	fmt.Println(hex.EncodeToString(checksum[:]))
	// Print out the execution time.
	fmt.Printf("took %f seconds\n", end.Sub(start).Seconds())
}
