package common

import (
	"fmt"
	"os"
)

func Require(b bool, err error) {
	if b {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Transform[T any, R any](slice []T, f func(t T) R) []R {
	r := make([]R, len(slice))
	for i, v := range slice {
		r[i] = f(v)
	}
	return r
}
