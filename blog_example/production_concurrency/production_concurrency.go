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

func loadWebsite(url string, contentChan chan []byte, errChan chan error) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("error loading: %v", r)
			errChan <- err
		}
	}()
	// Perform an HTTP GET to retrieve the website contents.
	val, err := http.Get(url)
	if err != nil {
		// We aren't interested in handling errors, so immediately exit the program if
		// an error is encountered
		errChan <- err
	}
	// Next, we read the entire contents of the response's body
	contents, err := ioutil.ReadAll(val.Body)
	if err != nil {
		errChan <- err
	}
	// After reading the body, we send it on a channel.
	contentChan <- contents
}

func getChecksum(urls []string) (string, error) {
	// Allocate a slice to store the contents of all of the websites we are
	// interested in.
	allContents := []byte{}

	contentChan := make(chan []byte)
	errChan := make(chan error)
	for _, website := range urls {
		go loadWebsite(website, contentChan, errChan)
	}

	// Collect the contents together
	for range listOfWebsites {
		select {
		case contents := <-contentChan:
			allContents = append(allContents, contents...)
		case err := <-errChan:
			return "", err
		}
	}
	checksum := sha512.Sum512(allContents)
	return hex.EncodeToString(checksum[:]), nil
}

func main() {
	// Record the start time of the program so that we can figure out how fast
	// it runs.
	start := time.Now()

	checksum, err := getChecksum(listOfWebsites)
	// select over err and contentChan

	// All done, record the time again.
	end := time.Now()

	// Calculate a checksum of the responses
	// Print out a human-readable form of the checksum
	// Print out the execution time.
	fmt.Printf("took %f seconds\n", end.Sub(start).Seconds())

	if err != nil {
		panic(err)
	}
	fmt.Println(checksum)
}
