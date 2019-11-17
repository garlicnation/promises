package promise

import (
	"fmt"
	"time"
)

func ExamplePromise_chain() {
	p := New(func() int {
		fmt.Println("second")
		time.Sleep(50 * time.Millisecond)
		return 7
	})

	pByTwo := p.Then(func(x int) int {
		fmt.Println("fifth")
		time.Sleep(100 * time.Millisecond)
		return x * 2
	})

	pByFour := p.Then(func(x int) int {
		fmt.Println("third")
		time.Sleep(50 * time.Millisecond)
		return x * 4
	})

	var resOne, resTwo int
	fmt.Println("First this happens")
	err := pByTwo.Wait(resOne)
	if err != nil {
		panic(err)
	}
	fmt.Println("fourth")
	err = pByFour.Wait(&resTwo)
	if err != nil {
		panic(err)
	}
	fmt.Println("sixth")
}
