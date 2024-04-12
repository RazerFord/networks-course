package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

func main() {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint32(bs, math.MaxUint32)
	fmt.Println(bs)
}