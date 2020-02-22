package promise

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPromiseResolution(t *testing.T) {
	p := New(func() int {
		return 1
	})
	var resolved int
	err := p.Wait(&resolved)
	require.Nil(t, err, "The promise should resolve without error")
	require.Equal(t, 1, resolved, "The promise should return a bare integer properly")
}

func TestPromiseCreationFailsWithoutFunction(t *testing.T) {
	require.Panics(t, func() {
		_ = New(4)
	}, "New should fail if it's not provided a function")
}

func TestThenFailsWithoutFunction(t *testing.T) {
	p := New(func() int {
		return 1
	})
	require.Panics(t, func() {
		p.Then(4)
	}, "New should fail if it's not provided a function")
}

func TestPromiseResolutionBadReturnType(t *testing.T) {
	p := New(func() string {
		return "garlic"
	})
	var resolved int
	require.Panics(t, func() {
		p.Wait(&resolved)
	}, "Returing a string into an int is not allowed")
}

func TestPromiseResolutionVoidReturnError(t *testing.T) {
	p := New(func() {
	})
	var resolved int
	require.Panics(t, func() {
		p.Wait(&resolved)
	}, "A function that returns void cannot return an int")
}

func TestPromiseResolutionWrongArgumentType(t *testing.T) {
	require.Panics(t, func() {
		_ = New(func(_ int) {
		}, "sizzle")
	}, "A function that accepts a int cannot accept a string")
}

func TestPromiseResolutionChain(t *testing.T) {
	returnOne := New(func(x int) int {
		return x
	}, 7)

	multiplyByTwo := returnOne.Then(func(x int) int {
		return x*2 + 3
	})

	var result int
	err := multiplyByTwo.Wait(&result)
	require.Nil(t, err)
	require.Equal(t, 17, result)
}

func TestPromiseAll(t *testing.T) {
	returnSeven := New(func(x int) int {
		return x
	}, 7)

	returnEight := New(func(x int) int {
		return x
	}, 8)

	returnNine := New(func(x int) int {
		return x
	}, 9)

	returnTen := New(func(x int) int {
		return x
	}, 10)

	returnEleven := New(func(x int) int {
		return x
	}, 11)

	returnAll := All(returnSeven, returnEight, returnNine, returnTen, returnEleven)

	var seven, eight, nine, ten, eleven int
	err := returnAll.Wait(&seven, &eight, &nine, &ten, &eleven)
	require.Nil(t, err)
	require.Equal(t, 7, seven)
	require.Equal(t, 8, eight)
	require.Equal(t, 9, nine)
	require.Equal(t, 10, ten)
	require.Equal(t, 11, eleven)
}

func TestPromiseAllReturnIntoSlice(t *testing.T) {
	returnSeven := New(func(x int) int {
		return x
	}, 7)

	returnEight := New(func(x int) int {
		return x
	}, 8)

	returnNine := New(func(x int) int {
		return x
	}, 9)

	returnTen := New(func(x int) int {
		return x
	}, 10)

	returnEleven := New(func(x int) int {
		return x
	}, 11)

	promises := []*Promise{returnSeven, returnEight, returnNine, returnTen, returnEleven}

	returnAll := All(promises...)

	returnSlice := returnAll.Then(func(vals ...int) []int {
		return vals
	})

	values := []int{}

	err := returnSlice.Wait(&values)
	require.Nil(t, err)
	require.EqualValues(t, []int{7, 8, 9, 10, 11}, values)
}

func TestPromiseAllDirectReturnIntoSlice(t *testing.T) {
	returnSeven := New(func(x int) int {
		return x
	}, 7)

	returnEight := New(func(x int) int {
		return x
	}, 8)

	returnNine := New(func(x int) int {
		return x
	}, 9)

	returnTen := New(func(x int) int {
		return x
	}, 10)

	returnEleven := New(func(x int) int {
		return x
	}, 11)

	promises := []*Promise{returnSeven, returnEight, returnNine, returnTen, returnEleven}

	returnAll := All(promises...)

	values := []int{}

	err := returnAll.Wait(&values)
	require.Nil(t, err)
	require.EqualValues(t, []int{7, 8, 9, 10, 11}, values)
}

func TestAllReturnsIfAnyFails(t *testing.T) {
	neverendingPromise := New(func() {
		time.Sleep(100000 * time.Second)
	})

	failingPromise := New(func() {
		panic("Failed!")
	})

	result := All(neverendingPromise, failingPromise)
	err := result.Wait()
	require.NotNil(t, err)
}

func BenchmarkPromiseAllReturnIntoSlice(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		returnSeven := New(func(x int) int {
			return x
		}, 7)

		returnEight := New(func(x int) int {
			return x
		}, 8)

		returnNine := New(func(x int) int {
			return x
		}, 9)

		returnTen := New(func(x int) int {
			return x
		}, 10)

		returnEleven := New(func(x int) int {
			return x
		}, 11)

		promises := []*Promise{returnSeven, returnEight, returnNine, returnTen, returnEleven}

		returnAll := All(promises...)

		returnSlice := returnAll.Then(func(vals ...int) []int {
			return vals
		})

		values := []int{}

		err := returnSlice.Wait(&values)
		require.Nil(b, err)
		require.EqualValues(b, []int{7, 8, 9, 10, 11}, values)
	}
}

func BenchmarkSyncSlicesWithChannels(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		values := []int{}
		valueChan := make(chan int)

		go func() {
			valueChan <- 7
		}()

		go func() {
			valueChan <- 8
		}()

		go func() {
			valueChan <- 9
		}()

		go func() {
			valueChan <- 10
		}()

		go func() {
			valueChan <- 11
		}()

		for i := 0; i < 5; i++ {
			values = append(values, <-valueChan)
		}

		require.ElementsMatch(b, []int{7, 8, 9, 10, 11}, values)
	}
}

func TestErrorReturnExitsEarly(t *testing.T) {
	instantError := New(func() error {
		return errors.New("error")
	})
	blocker := make(chan struct{})
	waitForever := New(func() {
		for range blocker {
		}
	})
	all := All(instantError, waitForever)
	err := all.Wait()
	close(blocker)
	require.Error(t, err)
}

func TestPromiseRaceSucceedsIfOneSucceeds(t *testing.T) {
	sleepThenPanic := func() string {
		time.Sleep(100 * time.Millisecond)
		panic("failed")
		return ""
	}

	sleepThenErr := func() (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "", fmt.Errorf("err")
	}

	success := func() string {
		return "success"
	}

	result := Race(New(sleepThenErr), New(sleepThenPanic), New(success))
	var retval string
	err := result.Wait(&retval)
	require.NoError(t, err)
	require.Equal(t, "success", retval)
}

func TestPromiseRaceFailsIfOneErrors(t *testing.T) {
	sleepThenPanic := func() string {
		time.Sleep(100 * time.Millisecond)
		panic("failed")
		return ""
	}

	returnError := func() (string, error) {
		return "", fmt.Errorf("err")
	}

	sleepThenSuccess := func() string {
		time.Sleep(100 * time.Millisecond)
		return "success"
	}

	result := Race(New(returnError), New(sleepThenPanic), New(sleepThenSuccess))
	var retval string
	err := result.Wait(&retval)
	require.Error(t, err)
	require.Contains(t, err.Error(), "err")
	require.Equal(t, "", retval)
}

func TestPromiseRaceFailsIfOnePanics(t *testing.T) {
	justPanic := func() string {
		panic("failed")
		return ""
	}

	sleepThenError := func() (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "", fmt.Errorf("err")
	}

	sleepThenSuccess := func() string {
		time.Sleep(100 * time.Millisecond)
		return "success"
	}

	result := Race(New(sleepThenError), New(justPanic), New(sleepThenSuccess))
	var retval string
	err := result.Wait(&retval)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed")
	require.Equal(t, "", retval)
}
