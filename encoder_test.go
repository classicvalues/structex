/*
Copyright 2021 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.

IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE
USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package structex

import (
	"encoding/binary"
	"fmt"
	"testing"
)

const (
	BufferSize = 512
)

type testWriter struct {
	bytes  [BufferSize]byte
	nbytes int
}

func (tw *testWriter) WriteByte(b byte) error {
	if tw.nbytes >= BufferSize {
		return fmt.Errorf("Byte buffer overflow")
	}

	tw.bytes[tw.nbytes] = b
	tw.nbytes++

	return nil
}

func (tw *testWriter) getByte(i int) byte {
	return tw.bytes[i]
}

func (tw *testWriter) getBytes(start, end int) []byte {
	b := make([]byte, end-start+1)
	for idx := range b {
		b[idx] = tw.getByte(start + idx)
	}
	return b
}

func (tw *testWriter) getSize() int {
	return tw.nbytes
}

func packAndTest(t *testing.T, s interface{}, testFunc func(t *testing.T, tw *testWriter)) {
	var tw = &testWriter{}
	if err := Encode(tw, s); err != nil {
		t.Error(err)
	}

	testFunc(t, tw)
}

func TestSimpleEncoder(t *testing.T) {
	s := struct {
		A uint16
		B uint16
	}{
		0x0001, 0xFFEE,
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 0x01 {
			t.Errorf("Simple pack failure byte 0: Expected: %#02x Actual: %#02x", 0x01, tw.getByte(0))
		}
		if tw.getByte(1) != 0x00 {
			t.Errorf("Simple pack failure byte 1: Expected: %#02x Actual: %#02x", 0x00, tw.getByte(1))
		}
		if tw.getByte(2) != 0xEE {
			t.Errorf("Simple pack failure byte 2: Expected: %#02x Actual: %#02x", 0xEE, tw.getByte(2))
		}
		if tw.getByte(3) != 0xFF {
			t.Errorf("Simple pack failure byte 3: Expected: %#02x Actual: %#02x", 0xFF, tw.getByte(3))
		}
	})
}

func TestEndianEncoder(t *testing.T) {
	s := struct {
		Big16    uint16 `structex:"big"`
		Little16 uint16
		Big32    uint32 `structex:"big"`
		Little32 uint32
		Big64    uint64 `structex:"big"`
		Little64 uint64

		Big32Forced    uint32 `structex:"big"`
		Little32Forced uint32 `structex:"little"`
	}{
		0x0123, 0x0123,
		0x01234567, 0x01234567,
		0x0123456789ABCDEF, 0x0123456789ABCDEF,
		0x89ABCDEF, 0x89ABCDEF,
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		// uint16
		{

			big16 := binary.BigEndian.Uint16(tw.getBytes(0, 1))
			little16 := binary.LittleEndian.Uint16(tw.getBytes(2, 3))

			if big16 != s.Big16 {
				t.Errorf("Invalid big-endian value for 16-bit field: Expected: %#04x Actual: %#04x", s.Big16, big16)
			}
			if little16 != s.Little16 {
				t.Errorf("Invalid little-endian value for 16-bit field: Expected: %#04x Actual: %#04x", s.Little16, little16)
			}
		}

		// uint32
		{
			big32 := binary.BigEndian.Uint32(tw.getBytes(4, 7))
			little32 := binary.LittleEndian.Uint32(tw.getBytes(8, 11))

			if big32 != s.Big32 {
				t.Errorf("Invalid big-endian value for 32-bit field: Expected: %#08x Actual: %#08x", s.Big32, big32)
			}
			if little32 != s.Little32 {
				t.Errorf("Invalid little-endian value for 32-bit field: Expected: %#08x Actual: %#08x", s.Little32, little32)
			}
		}

		// uint64
		{
			big64 := binary.BigEndian.Uint64(tw.getBytes(12, 19))
			little64 := binary.LittleEndian.Uint64(tw.getBytes(20, 27))

			if big64 != s.Big64 {
				t.Errorf("Invalid big-endian value for 64-bit field: Expected: %#08x Actual: %#08x", s.Big64, big64)
			}
			if little64 != s.Little64 {
				t.Errorf("Invalid little-endian value for 64-bit field: Expected: %#08x Actual: %#08x", s.Little64, little64)
			}
		}

		// uint32 - forced endian formats
		{
			big32 := binary.BigEndian.Uint32(tw.getBytes(28, 31))
			little32 := binary.LittleEndian.Uint32(tw.getBytes(32, 35))

			if big32 != s.Big32Forced {
				t.Errorf("Invalid FORCED big-endian value for 32-bit field: Expected: %#08x Actual: %#08x", s.Big32Forced, big32)
			}

			if little32 != s.Little32Forced {
				t.Errorf("Invalid FORCED little-endian value for 32-bit field: Expected: %#08x Actual: %#08x", s.Little32Forced, little32)
			}
		}
	})
}

func TestBitfieldEncoder(t *testing.T) {

	s := struct {
		A int `bitfield:"3"`
		B int `bitfield:"4"`
		C int `bitfield:"1"`
		D int `bitfield:"12"`
		E int `bitfield:"4"`
	}{
		0x7, 0x8, 0x1, 0x0FFF, 0x1,
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 0xC7 {
			t.Errorf("Invalid bitfield: Expected: %#02x Actual: %#02x", 0xC7, tw.getByte(0))
		}
		if tw.getByte(1) != 0xFF {
			t.Errorf("Invalid bitfield: Expected: %#02x Actual: %#02x", 0xFF, tw.getByte(1))
		}
		if tw.getByte(2) != 0x1F {
			t.Errorf("Invalid bitfield: Expected: %#02x Actual: %#02x", 0x1F, tw.getByte(2))
		}
	})
}

func TestNestingEncoder(t *testing.T) {
	type Nested struct {
		M uint8
		N uint8
	}

	type S struct {
		A uint8
		B uint8
		C Nested
	}

	s := S{
		A: 0x00,
		B: 0x01,
		C: Nested{
			M: 0x02,
			N: 0x03,
		},
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		for i := 0; i < 4; i++ {
			if tw.getByte(i) != uint8(i) {
				t.Errorf("Invalid byte at offset %d: Expected: %#02x Actual: %#02x", i, i, tw.getByte(i))
			}
		}
	})

}

func TestByteArrayEncoder(t *testing.T) {
	s := struct {
		Size uint8
		Data []byte
	}{
		Size: 4,
		Data: []byte{0x00, 0x01, 0x02, 0x03},
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 4 {
			t.Errorf("Size Encoding Incorrect: Expected: %d Actual: %d", 4, tw.getByte(0))
		}
		for i := 0; i < 4; i++ {
			if tw.getByte(i+1) != byte(i) {
				t.Errorf("Byte Index %d Incorrect: Expected: %#02x Actual: %#02x", i, i, tw.getByte(i+1))
			}
		}
	})
}
func TestArrayEncoder(t *testing.T) {
	type T struct {
		A uint8
		B uint8
	}

	s := struct {
		Count uint8 `countOf:"Ts"`
		Size  uint8 `sizeOf:"Ts"`
		Ts    [2]T
	}{
		Count: 0x00,
		Size:  0x00,
		Ts: [2]T{
			{A: 0x01, B: 0x02},
			{A: 0x03, B: 0x04},
		},
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 2 {
			t.Errorf("Invalid countOf: Expected: %d Actual: %d", 2, tw.getByte(0))
		}
		if tw.getByte(1) != 4 {
			t.Errorf("Invalid sizeOf: Expected: %d Actual: %d", 4, tw.getByte(1))
		}

		expected := []uint8{0x01, 0x02, 0x03, 0x04}
		actual := []uint8{tw.getByte(2), tw.getByte(3), tw.getByte(4), tw.getByte(5)}

		for i := 0; i < 4; i++ {
			if expected[i] != actual[i] {
				t.Errorf("Invalid array pack: Index: %d Expected: %#02x Actual: %#02x", i, expected[i], actual[i])
			}
		}
	})
}

func TestArrayLittleEndian16Encoder(t *testing.T) {
	s := struct {
		Count uint8 `countOf:"Ts"`
		Size  uint8 `sizeOf:"Ts"`
		Ts    [2]uint16 `little:""`
	}{
		Count: 0x00,
		Size:  0x00,
		Ts: [2]uint16{0x0102, 0x0304},
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 2 {
			t.Errorf("Invalid countOf: Expected: %d Actual: %d", 2, tw.getByte(0))
		}
		if tw.getByte(1) != 4 {
			t.Errorf("Invalid sizeOf: Expected: %d Actual: %d", 4, tw.getByte(1))
		}

		expected := []uint8{0x02, 0x01, 0x04, 0x03}
		actual := []uint8{tw.getByte(2), tw.getByte(3), tw.getByte(4), tw.getByte(5)}

		for i := 0; i < 4; i++ {
			if expected[i] != actual[i] {
				t.Errorf("Invalid array pack: Index: %d Expected: %#02x Actual: %#02x", i, expected[i], actual[i])
			}
		}
	})
}

func TestArrayBigEndian16Encoder(t *testing.T) {
	s := struct {
		Count uint8 `countOf:"Ts"`
		Size  uint8 `sizeOf:"Ts"`
		Ts    [2]uint16 `big:""`
	}{
		Count: 0x00,
		Size:  0x00,
		Ts: [2]uint16{0x0102, 0x0304},
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 2 {
			t.Errorf("Invalid countOf: Expected: %d Actual: %d", 2, tw.getByte(0))
		}
		if tw.getByte(1) != 4 {
			t.Errorf("Invalid sizeOf: Expected: %d Actual: %d", 4, tw.getByte(1))
		}

		expected := []uint8{0x01, 0x02, 0x03, 0x04}
		actual := []uint8{tw.getByte(2), tw.getByte(3), tw.getByte(4), tw.getByte(5)}

		for i := 0; i < 4; i++ {
			if expected[i] != actual[i] {
				t.Errorf("Invalid array pack: Index: %d Expected: %#02x Actual: %#02x", i, expected[i], actual[i])
			}
		}
	})
}

func TestSliceEncoder(t *testing.T) {

	ts := [6]uint8{0xA, 0xB, 0xC, 0xD, 0xE, 0xF}

	s := struct {
		Count uint8 `countOf:"Ts"`
		Size  uint8 `sizeOf:"Ts"`
		Ts    []uint8
	}{
		Count: 0,
		Size:  0,
		Ts:    ts[2:4],
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 2 {
			t.Errorf("Invalid countOf: Expected: %d Actual: %d", 2, tw.getByte(0))
		}

		if tw.getByte(1) != 2 {
			t.Errorf("Invalid sizeOf: Expected: %d Actual: %d", 2, tw.getByte(1))
		}

		// Check the slice contents
		expected := []uint8{0x0C, 0x0D}
		actual := []uint8{tw.getByte(2), tw.getByte(3)}

		for i := 0; i < len(expected); i++ {
			if expected[i] != actual[i] {
				t.Errorf("Invalid slice pack: Index: %d Expected: %#02x Actual: %#02x", i, expected[i], actual[i])
			}
		}
	})
}

func TestArrayTruncate(t *testing.T) {
	s := struct {
		Size  uint32 `sizeOf:"Array"`
		Array [BufferSize - 4]byte
	}{
		4,
		[BufferSize - 4]byte{0x00, 0x01, 0x02, 0x03, 0xFF /*Shouldn't be decoded*/},
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getSize() != 8 {
			t.Errorf("Invalid size of encoded buffer: Expected: %d Actual: %d", 8, tw.getSize())
		}
		if tw.getByte(0) != 4 {
			t.Errorf("Invalid sizeOf: Expected: %d Actual: %d", 4, tw.getByte(0))
		}
		for i := 0; i < 4; i++ {
			if tw.getByte(4+i) != byte(i) {
				t.Errorf("Invalid array byte: Expected: %#02x Actual: %#02x", i, tw.getByte(4+i))
			}
		}
	})
}

func TestAlignment(t *testing.T) {
	s := struct {
		Pad     uint8
		Aligned uint32 `align:"4"`
	}{
		0x00, 0xFF,
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getSize() != 8 {
			t.Errorf("Invalid size of encoded buffer: Expected: %d Actual: %d", 8, tw.getSize())
		}

		if tw.getByte(4) != 0xFF {
			t.Errorf("Invalid aligned field: Expected %#02x Actual: %#02x", 0xFF, tw.getByte(4))
		}
	})
}

func TestBoolEncoder(t *testing.T) {
	s := struct {
		IsD  bool  `bitfield:"1"`
		ValC uint8 `bitfield:"2"`
		IsC  bool  `bitfield:"1"`
		IsB  bool  `bitfield:"1"`
		ValB uint8 `bitfield:"2"`
		IsA  bool  `bitfield:"1"`
		ValD uint8
	}{
		true, 0b10, true, false, 0b01, true, 0x12,
	}

	packAndTest(t, s, func(t *testing.T, tw *testWriter) {
		if tw.getByte(0) != 0b10101101 {
			t.Errorf("Invalid bitfield: Expected: %d Actual: %d", 0b10101101, tw.getByte(0))
		}
		if tw.getByte(1) != 0x12 {
			t.Errorf("Invalid bitfield: Expected: %d Actual: %d", 0x12, tw.getByte(1))
		}
	})
}
