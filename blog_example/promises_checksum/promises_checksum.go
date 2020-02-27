package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	promise "github.com/garlicnation/promises/v2"
)

var listOfWebsites = []string{
	"https://google.com",
	"https://facebook.com",
	"https://yahoo.com",
	"https://bing.com",
	"https://github.com",
}

func main() {
	// Record the start time of the program so that we can figure out how fast
	// it runs.
	start := time.Now()

	allFetches := []*promise.Promise{}

	for _, website := range listOfWebsites {
		promise := promise.New(http.Get, website).Then(
			func(resp *http.Response) io.Reader {
				return resp.Body
			}).Then(ioutil.ReadAll)
		allFetches = append(allFetches, promise)
	}

	var responses [][]byte

	err := promise.All(allFetches...).Wait(&responses)
	if err != nil {
		fmt.Printf("Error in fetch: %+v", err)
		panic(err)
	}

	// All done, record the time again.
	end := time.Now()

	contents := []byte{}
	for _, response := range responses {
		contents = append(contents, response...)
	}

	// Calculate a checksum of the responses
	checksum := sha512.Sum512(contents)
	// Print out a human-readable form of the checksum
	fmt.Println(hex.EncodeToString(checksum[:]))
	// Print out the execution time.
	fmt.Printf("took %f seconds\n", end.Sub(start).Seconds())
}
