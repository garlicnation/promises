package main

import (
	"testing"

	promise "github.com/garlicnation/promises/v2"
)

var values []int

func BenchPromiseReturnIntoSlice(b *testing.B) {

	for i := 0; i < b.N; i++ {

		returnSeven := promise.New(func(x int) int {
			return x
		}, 7)

		returnEight := promise.New(func(x int) int {
			return x
		}, 8)

		returnNine := promise.New(func(x int) int {
			return x
		}, 9)

		returnTen := promise.New(func(x int) int {
			return x
		}, 10)

		returnEleven := promise.New(func(x int) int {
			return x
		}, 11)

		promises := []*promise.Promise{returnSeven, returnEight, returnNine, returnTen, returnEleven}

		returnAll := promise.All(promises...)

		returnSlice := returnAll.Then(func(vals ...int) []int {
			return vals
		})

		err := returnSlice.Wait(&values)
		if err != nil {
			b.Fatal(err)
		}
	}
}
