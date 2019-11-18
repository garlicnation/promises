/*
Promises is a library that builds something similar to JS style promises, or Futures(as seen in Java and other languages) for golang.

Promises is type-safe at runtime, and within an order of magnitude of performance of a solution built with pure channels and goroutines.

For a more thorough introduction to the library, please check out https://github.com/garlicnation/promises/blob/master/blog_example/WHY.md

Examples

Single promise:
	p := promise.New(func() int {
		return 1
	})
	var resolved int
	err := p.Wait(&resolved)

Chained Promise:
	p := promise.New(func() int {
		return 1
	})
	timesTwo := p.Then(func(x int) int {
		return x*2
	})
	var resolved int
	err := timesTwo.Wait(&resolved)

Promise.all:
	p := promise.New(func() int {
		return 1
	})
	timesTwo := p.Then(func(x int) int {
		return x * 2
	})
	plusFour := timesTwo.Then(func(x int) int{
		return x + 4
	})

	all := promise.All(p, timesTwo, plusFour)
	results := []int{}
	err := all.Wait(&results)

Error handling:
	p := promise.New(http.Get, "http://example.com")
	// Do other work while Get processes in a goroutine...
	var r *http.Response
	err := p.Wait(&r)
	// Errors are swallowed by promise and returned by wait.

*/
package promise
