# Why Build Promises?

I've worked pretty extensively with go for the last few years, and have mostly enjoyed the experience. I find that go projects are often well-structured, that the go's interface concept makes it easy to build lightly coupled code, and working with the tooling around go is a joy.

However, one of my biggest complaints about go is how hard some basic concurrency patterns are. goroutines are light and easy to spawn and channels are generic and powerful enough to build many kinds of concurrent algorithms, but there are many edge cases in using these tools, and there is a lot of repetition in handling them correctly.

To illustrate what I mean, let's walk through an example that's close to a pattern that often occurs in real-world go programming.

To start, let's build a basic go utility that calculates a checksum of 5 websites. If the checksum changes, we know that one of the five websites has changed. We'll explore its performance characteristics on our journey to make the code production quality and optimized. After seeing the complexity of the full production code, we'll simplify it with promises without reducing its capabilities at all.

### Prerequisites

This document assumes basic knowledge of the go programming language, HTTP GET requests, JavaScript Promises, and the concept of a checksum. If you feel like you need a refresher on any of the above, I've provided some links that I found educational on those subjects:
- [A Tour of Go](https://tour.golang.org/welcome/1)
- [What is a Checksum?](https://www.lifewire.com/what-does-checksum-mean-2625825)
- [HTTP Basics](https://www.ntu.edu.sg/home/ehchua/programming/webprogramming/HTTP_Basics.html)
- [What is a Promise?](https://medium.com/javascript-scene/master-the-javascript-interview-what-is-a-promise-27fc71e77261)

### First Draft

We'd like to build a tool that determines if any of a list of major websites changes. Our basic approach will be to make requests to the list of urls, then use the go standard library to compute a checksum.

[embedmd]:# (serial_checksum/serial_checksum.go)
```go
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
```

Looks pretty good for a first pass! This program successfully meets the requirements we defined at the beginning, but can it be improved?

First off, how long does it take to run?

```bash
user@laptop:~/workspace/promises/blog_example/serial_checksum$ ./serial_checksum 
a5409feb6fcf435f965a51914a5411a0521d055b311f46bc9612e2d8df6f2c4c89b2a697e7ef7c1ead849ce8bcdb2b8f0175ab9cbc43a78e582143649a56c4f7
took 1.487321 seconds
```

It works! But... 1.48 seconds? Surely we can do better than that. This seems like a perfect use case for some tools that are built into go, channels and goroutines. Let's see what it looks like to add these in.

### Basic Concurrency

[embedmd]:# (basic_concurrency/basic_concurrency.go /func main/ $)
```go
func main() {
	// Allocate a slice to store the contents of all of the websites we are
	// interested in.
	allContents := []byte{}
	// Record the start time of the program so that we can figure out how fast
	// it runs.
	start := time.Now()

	contentChan := make(chan []byte)

	for _, website := range listOfWebsites {
		go func(w string) {
			// Perform an HTTP GET to retrieve the website contents.
			val, err := http.Get(w)
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
			// After reading the body, we send it on a channel.
			contentChan <- contents
		}(website)
	}

	// Collect the contents together
	for range listOfWebsites {
		contents := <-contentChan
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
```

I'll dive into the changes I made, but first, did we make things faster?

```
user@laptop:~/workspace/promises/blog_example/basic_concurrency$ ./basic_concurrency 
b4b52abb71f08ab6cbb5725acf93b75c0e091fbc66ec3567bc55b56157141665031ccbeda623eee9d0687304e35d0ed004fb1b27dcddb1450f4cb3b75c37021a
took 0.564650 seconds
```

Wow! Applying just a bit of concurrency sped things up by close to 50%! It's not normally so easy to get this kind of speed up, but was it worth the extra complexity we've introduced? This program now has a number of maintenance pitfalls not present in the original.

Take this line:
```go
contentChan := make(chan []byte)
```

`contentChan` is unbuffered, which means that if we try to read from it before any of the goroutines have written something, we will block forever. It's very important that the loop we added comes after launching goroutines.

What else?

Well, the `panic` calls are certainly more problematic than they were before. Any time a `panic` occurs in go, if it isn't recovered, the entire application crashes. This is fine for our little utility, but is unacceptable in larger applications that might be serving many concurrent requests.

### Production ready concurrency
[embedmd]:# (production_concurrency/production_concurrency.go /func loadWebsite/ $)
```go
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
```

Alright, this resembles something that we might want to embed in a larger application. We've now added the necessary panic handler to our goroutine, and added a channel for errors to enable early return.

```
user@laptop:~/workspace/promises/blog_example/production_concurrency$ ./production_concurrency 
took 0.702972 seconds
a6317fe59e5752210ee3194b61fc6053cc6f4b1a1990fba64aa636560e1e16462256504baca5b99b698b327fcb15c02939f2c7cbe5257af889cfae6cb97cf947

```

With the new error channel, we've introduced a possibility for deadlocks, but we avoid that by using the `select` statement. This code works great, but is starting to get more and more complex, even though we're not doing that complex of a workflow. This kind of code is what inspired me to write promises. In a moment, you'll see how much simpler things can be.


### Concurrency with promises
Using promises, we can reduce the entire critical section to just a few lines, and remove all channels and goroutines.

This below snippet replaces all of the coordination code above! Including exiting early for an error, and also has some features that
the above code doesn't have, like fan out for return values. Additionally, it fixes a bug in the above code where the checksum will vary based on the order in which we receive our HTTP responses. This code will compute a stable checksum if the HTTP responses are identical.

```
	for _, website := range listOfWebsites {
		promise := promises.New(http.Get, website).Then(
			func(resp *http.Response) io.Reader {
				return resp.Body
			}).Then(ioutil.ReadAll)
		allFetches = append(allFetches, promise)
	}
```

Below, see how this fits into an overall application:

[embedmd]:# (promises_checksum/promises_checksum.go /func main/ $)
```go
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
```

Not just that, but it's possible to compose these promises further! It's safe to call `.Then` as many times as you'd like on a promise, meaning that any kind of concurrent workflow you can imagine is simple to implement. Not just that, but any library that follows the go convention of returning `value1, value2, ... value2, error` automatically has error propegation with promises, meaning you can chain large numbers of operations without any `if err != nil` boilerplate!

As far as rutime goes, this last version is just as fast as the version built with channels!

```
user@laptop:~/workspace/promises/blog_example/promises_checksum$ ./promises_checksum 
63d6b5155ca1e7cc36f7c97e9c26b0a1b65277be3bc4a088722175f03bf01495e7e2803ea9085c686dc0adaae5ef95ff3f7528b0b4836fbd30216f214a9a600d
took 0.677518 seconds
```

I hope you enjoy using this library as much as I enjoyed creating it. Please report any issues to the issue page, and PRs are always welcome!

Thanks,
AJ
