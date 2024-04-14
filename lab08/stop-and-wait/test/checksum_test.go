package checksum_test

import (
	"math"
	"stop-and-wait/internal/network/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidCheckSum(t *testing.T) {
	data := []struct {
		data     []byte
		checksum uint16
	}{
		{[]byte(""), math.MaxUint16},
		{[]byte{1}, math.MaxUint16 ^ 1},
		{[]byte{0, 0, 0, 0, 0, 1}, math.MaxUint16 ^ 1},
		{[]byte{0, 0, 0, 0, 1, 0}, math.MaxUint16 ^ (1 << 8)},
		{[]byte("Hello, world!"), 57236},
	}

	for _, d := range data {
		assert.Equal(t, d.checksum, common.ToChecksum(d.data))
	}

	for _, d := range data {
		assert.True(t, common.CheckChecksum(d.data, d.checksum))
	}
}

func Test_InvalidCheckSum(t *testing.T) {
	data := []struct {
		data     []byte
		checksum uint16
	}{
		{[]byte(""), 1},
		{[]byte{1}, math.MaxUint16},
		{[]byte{1, 0, 0, 0}, 124},
		{[]byte{1, 1, 0}, 17},
		{[]byte("Hello, world!"), 6},
	}

	for _, d := range data {
		assert.NotEqual(t, d.checksum, common.ToChecksum(d.data))
	}

	for _, d := range data {
		assert.False(t, common.CheckChecksum(d.data, d.checksum), "%v %v", d.checksum, common.ToChecksum(d.data))
	}
}
