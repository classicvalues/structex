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
	"bytes"
	"fmt"
	"math/bits"
	"testing"
)

const (
	bufferSize = 512
)

type testReader struct {
	bytes  [bufferSize]byte
	offset int
	nbytes int
}

func newReader(b []byte) *testReader {
	var tr = new(testReader)
	copy(tr.bytes[:], b)
	tr.offset = 0
	tr.nbytes = len(b)
	return tr
}

func (tr *testReader) ReadByte() (byte, error) {
	if tr.offset == tr.nbytes {
		return 0, fmt.Errorf("Buffer underrun")
	}

	b := tr.bytes[tr.offset]
	tr.offset++
	return b, nil
}

func unpackAndTest(t *testing.T, s interface{}, tr *testReader, testFunc func(t *testing.T, s interface{})) {
	if err := Decode(tr, s); err != nil {
		t.Error(err)
	}

	testFunc(t, s)
}

func TestBasicDecoder(t *testing.T) {
	type ts struct {
		A uint8
		B uint16
		C uint32
		D uint64
	}

	var s = new(ts)

	var tr = newReader([]byte{
		0x00,
		0x01, 0x00,
		0x02, 0x00, 0x00, 0x00,
		0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)
		if s.A != 0 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x00, s.A)
		}
		if s.B != 0x0001 {
			t.Errorf("Test Value Incorrect: Expected: %#04x Actual: %#04x", 0x0001, s.B)
		}
	})
}

func TestEndianDecoder(t *testing.T) {
	type ts struct {
		Big16    uint16 `structex:"big"`
		Little16 uint16
		Big32    uint32 `structex:"big"`
		Little32 uint32
		Big64    uint64 `structex:"big"`
		Little64 uint64

		Big32Forced    uint32 `structex:"big"`
		Little32Forced uint32 `structex:"little"`
	}

	var s = new(ts)
	var tr = newReader([]byte{
		0x01, 0x23,
		0x01, 0x23,
		0x01, 0x23, 0x45, 0x67,
		0x01, 0x23, 0x45, 0x67,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
		0x89, 0xAB, 0xCD, 0xEF,
		0x89, 0xAB, 0xCD, 0xEF})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if s.Big16 != 0x0123 {
			t.Errorf("Invalid big-endian 16-bit value: Expected: %#04x Actual: %#04x", 0x0123, s.Big16)
		}
		if s.Little16 != bits.ReverseBytes16(0x0123) {
			t.Errorf("Invalid little-endian 16-bit value: Expected: %#04x Actual: %#04x", bits.ReverseBytes16(0x0123), s.Little16)
		}
		if s.Big32 != 0x01234567 {
			t.Errorf("Invalid big-endian 32-bit value: Expected: %#08x Actual: %#08x", 0x01234567, s.Big32)
		}
		if s.Little32 != bits.ReverseBytes32(0x01234567) {
			t.Errorf("Invalid little-endian 32-bit value: Expected: %#08x Actual: %#08x", bits.ReverseBytes32(0x01234567), s.Little32)
		}
		if s.Big64 != 0x0123456789ABCDEF {
			t.Errorf("Invalid big-endian 64-bit value: Expected: %#016x Actual: %#016x", 0x0123456789ABCDEF, s.Big64)
		}
		if s.Little64 != bits.ReverseBytes64(0x0123456789ABCDEF) {
			t.Errorf("Invalid little-endian 64-bit value: Expected: %#016x Actual: %#016x", bits.ReverseBytes64(0x0123456789ABCDEF), s.Little64)
		}
		if s.Big32Forced != 0x89ABCDEF {
			t.Errorf("Invalid FORCED big-endian 32-bit value: Expected %#08x Actual: %08x", 0x89ABCDEF, s.Big32Forced)
		}
		if s.Little32Forced != bits.ReverseBytes32(0x89ABCDEF) {
			t.Errorf("Invalid FORCED little-endian 32-bit value: Expected %#08x Actual: %08x", bits.ReverseBytes32(0x89ABCDEF), s.Little32Forced)
		}
	})

}
func TestBitfieldDecoder(t *testing.T) {

	type ts struct {
		A int `bitfield:"3"`
		B int `bitfield:"4"`
		C int `bitfield:"1"`
		D int `bitfield:"12"`
		E int `bitfield:"4"`
	}

	var s = new(ts)

	var tr = newReader([]byte{0xC7, 0xFF, 0x1F})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if s.A != 0x07 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x07, s.A)
		}
		if s.B != 0x08 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x08, s.B)
		}
		if s.C != 0x01 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x08, s.C)
		}
		if s.D != 0x0FFF {
			t.Errorf("Test Value Incorrect: Expected %#03x Actual: %#03x", 0xFFF, s.D)
		}
		if s.E != 0x1 {
			t.Errorf("Test Value Incorrect: Expected: %#x Actual: %#x", 0x1, s.E)
		}
	})
}

func TestNestedDecoder(t *testing.T) {

	type ns struct {
		M int `bitfield:"3"`
		N int `bitfield:"4"`
		O int `bitfield:"1"`
	}

	type ts struct {
		A uint8
		B uint8
		C ns
	}

	var s = new(ts)

	var tr = newReader([]byte{0x01, 0x02, 0xC7})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if s.A != 0x01 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x01, s.A)
		}
		if s.B != 0x02 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x02, s.B)
		}
		if s.C.M != 0x07 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x07, s.C.M)
		}
		if s.C.N != 0x08 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x08, s.C.N)
		}
		if s.C.O != 0x01 {
			t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x08, s.C.O)
		}
	})
}

func TestArrayDecoder(t *testing.T) {

	type as struct {
		A uint8 `bitfield:"4"`
		B uint8 `bitfield:"4"`
	}

	type ts struct {
		Count uint8 `countOf:"Cs"`
		Size  uint8 `sizeOf:"Ss"`
		Cs    []as
		Ss    []as
		A     [1]byte
	}

	var s = new(ts)

	var tr = newReader([]byte{2, 2, 0x11, 0x22, 0x33, 0x44, 0x55})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if s.Count != 2 {
			t.Errorf("Count Value Incorrect: Expected: %#02x Actual: %#02x", 2, s.Count)
		}
		if s.Size != 2 {
			t.Errorf("Size Value Incorrect: Expected: %#02x Actual: %#02x", 2, s.Size)
		}
		if s.Cs[0].A != 1 || s.Cs[0].B != 1 {
			val := s.Cs[0].A | (s.Cs[0].B << 4)
			t.Errorf("Array Value Incorrect: Expected: %#02x Actual: %#02x", 0x11, val)
		}
		if s.Cs[1].A != 2 || s.Cs[1].B != 2 {
			val := s.Cs[1].A | (s.Cs[1].B << 4)
			t.Errorf("Array Value Incorrect: Expected: %#02x Actual: %#02x", 0x22, val)
		}
		if s.Ss[0].A != 3 || s.Ss[0].B != 3 {
			val := s.Ss[0].A | (s.Ss[0].B << 4)
			t.Errorf("Array Value Incorrect: Expected: %#02x Actual: %#02x", 0x33, val)
		}
		if s.Ss[1].A != 4 || s.Ss[1].B != 4 {
			val := s.Ss[1].A | (s.Ss[1].B << 4)
			t.Errorf("Array Value Incorrect: Expected: %#02x Actual: %#02x", 0x44, val)
		}
		if s.A[0] != 0x55 {
			t.Errorf("Array Value Incorrect: Expected %#02x Actual: %#02x", 0x55, s.A[0])
		}
	})
}

func TestArrayLittleEndian16Decoder(t *testing.T) {
	type ts struct {
		Count uint8     `countOf:"Ts"`
		Size  uint8     `sizeOf:"Ts"`
		Ts    [2]uint16 `little:""`
	}

	var s = new(ts)

	var tr = newReader([]byte{2, 4, 0x34, 0x12, 0x78, 0x56})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if s.Count != 2 {
			t.Errorf("Count Value Incorrect: Expected: %#02x Actual: %#02x", 2, s.Count)
		}
		if s.Size != 4 {
			t.Errorf("Size Value Incorrect: Expected: %#02x Actual: %#02x", 4, s.Size)
		}
		if s.Ts[0] != 0x1234 {
			t.Errorf("Array Value Incorrect: Expected %#02x Actual: %#02x", 0x1234, s.Ts[0])
		}
		if s.Ts[1] != 0x5678 {
			t.Errorf("Array Value Incorrect: Expected %#02x Actual: %#02x", 0x5678, s.Ts[1])
		}
	})
}

func TestArrayBigEndian16Decoder(t *testing.T) {
	type ts struct {
		Count uint8     `countOf:"Ts"`
		Size  uint8     `sizeOf:"Ts"`
		Ts    [2]uint16 `big:""`
	}

	var s = new(ts)

	var tr = newReader([]byte{2, 4, 0x12, 0x34, 0x56, 0x78})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if s.Count != 2 {
			t.Errorf("Count Value Incorrect: Expected: %#02x Actual: %#02x", 2, s.Count)
		}
		if s.Size != 4 {
			t.Errorf("Size Value Incorrect: Expected: %#02x Actual: %#02x", 4, s.Size)
		}
		if s.Ts[0] != 0x1234 {
			t.Errorf("Array Value Incorrect: Expected %#02x Actual: %#02x", 0x1234, s.Ts[0])
		}
		if s.Ts[1] != 0x5678 {
			t.Errorf("Array Value Incorrect: Expected %#02x Actual: %#02x", 0x5678, s.Ts[1])
		}
	})
}

func TestHeaderStyleDecoder(t *testing.T) {

	type hdr struct {
		Count uint8 `countOf:"Details"`
	}

	type detail struct {
		Val uint8
	}

	type ts struct {
		Hdr     hdr
		Details []detail
	}

	s := new(ts)

	tr := newReader([]byte{4, 0, 1, 2, 3})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		s := i.(*ts)

		if s.Hdr.Count != 4 {
			t.Errorf("Count Value Incorrect: Expected: %d Actual: %d", 4, s.Hdr.Count)
		}

		if len(s.Details) != 4 {
			t.Errorf("Details Array Len Incorrect: Expected: %d Actual: %d", 4, len(s.Details))
		}

		for i := uint8(0); i < s.Hdr.Count; i++ {
			if s.Details[i].Val != i {
				t.Errorf("Detail Value Incorrect: Expected: %d Actual: %d", i, s.Details[i].Val)
			}
		}
	})
}

func TestBytesBuffer(t *testing.T) {

	type ts struct {
		A byte
		B byte
	}

	var s = new(ts)

	b := bytes.NewBuffer([]byte{0x01, 0x02})

	if err := DecodeByteBuffer(b, s); err != nil {
		t.Errorf("Unpack Buffer failed: Err: %v", err)
	}

	if s.A != 0x01 {
		t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x01, s.A)
	}

	if s.B != 0x02 {
		t.Errorf("Test Value Incorrect: Expected: %#02x Actual: %#02x", 0x02, s.B)
	}
}

func TestTruncate(t *testing.T) {
	type ts struct {
		A [1024]byte `truncate:""`
	}

	var s = new(ts)

	if err := DecodeByteBuffer(bytes.NewBuffer([]byte{0}), s); err != nil {
		t.Errorf("Truncate test failed: %s", err)
	}
}

func TestAlignmentDecoder(t *testing.T) {
	type ts struct {
		Pad     uint8
		Aligned uint32 `align:"4"`
	}

	var s = new(ts)

	tr := newReader([]byte{0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		s := i.(*ts)

		if s.Aligned != 0xFFFFFFFF {
			t.Errorf("Unexpected aligned parameter: Expected: %#08x Actual: %#08x", 0xFFFFFFFF, s.Aligned)
		}
	})
}

func TestBoolDecoder(t *testing.T) {
	type ts struct {
		IsD  bool  `bitfield:"1"`
		ValC uint8 `bitfield:"2"`
		IsC  bool  `bitfield:"1"`
		IsB  bool  `bitfield:"1"`
		ValB uint8 `bitfield:"2"`
		IsA  bool  `bitfield:"1"`
		ValD uint8

		ValF uint16 `bitfield:"2"`
		IsH  bool   `bitfield:"1"`
		IsG  bool   `bitfield:"1"`
		ValE uint16 `bitfield:"10"`
		IsF  bool   `bitfield:"1"`
		IsE  bool   `bitfield:"1"`
	}
	var s = new(ts)

	var tr = newReader([]byte{0b10101101, 0x12, 0b11110101, 0b10001111})

	unpackAndTest(t, s, tr, func(t *testing.T, i interface{}) {
		var s = i.(*ts)

		if !s.IsA {
			t.Errorf("IsA Value Incorrect: Expected: %v Actual: %v", true, s.IsA)
		}
		if s.ValB != 0b01 {
			t.Errorf("ValB Value Incorrect: Expected: %#02x Actual: %#02x", 0b01, s.ValB)
		}
		if s.IsB {
			t.Errorf("IsB Value Incorrect: Expected: %v Actual: %v", false, s.IsB)
		}
		if !s.IsC {
			t.Errorf("IsC Value Incorrect: Expected: %v Actual: %v", true, s.IsC)
		}
		if s.ValC != 0b10 {
			t.Errorf("ValC Value Incorrect: Expected: %#02x Actual: %#02x", 0b10, s.ValC)
		}
		if !s.IsD {
			t.Errorf("IsB Value Incorrect: Expected: %v Actual: %v", true, s.IsD)
		}
		if s.ValD != 0x12 {
			t.Errorf("ValC Value Incorrect: Expected: %#02x Actual: %#02x", 0x12, s.ValD)
		}
		if !s.IsE {
			t.Errorf("IsE Value Incorrect: Expected: %v Actual: %v", true, s.IsE)
		}
		if s.IsF {
			t.Errorf("IsF Value Incorrect: Expected: %v Actual: %v", false, s.IsF)
		}
		if s.ValE != 0b0011111111 {
			t.Errorf("ValE Value Incorrect: Expected: %#02x Actual: %#02x", 0x0011111111, s.ValE)
		}
		if s.IsG {
			t.Errorf("IsG Value Incorrect: Expected: %v Actual: %v", false, s.IsG)
		}
		if !s.IsH {
			t.Errorf("IsH Value Incorrect: Expected: %v Actual: %v", true, s.IsH)
		}
	})
}
