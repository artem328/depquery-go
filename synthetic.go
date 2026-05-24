package depquery

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/cespare/xxhash/v2"
)

func SyntheticID(parent uint64, entity []byte) uint64 {
	data := make([]byte, 8+len(entity))

	binary.BigEndian.PutUint64(data[:8], parent)
	copy(data[8:], entity)

	return xxhash.Sum64(data)
}

func IBytes[I ~int | ~uint](i I) []byte {
	switch unsafe.Sizeof(0) {
	case 4:
		return I32Bytes(uint32(i))
	case 8:
		fallthrough
	default:
		return I64Bytes(uint64(i))
	}
}

func I64Bytes[I ~int64 | ~uint64](i I) []byte {
	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, uint64(i))

	return buf
}

func I32Bytes[I ~int32 | ~uint32](i I) []byte {
	buf := make([]byte, 4)

	binary.BigEndian.PutUint32(buf, uint32(i))

	return buf
}

func I16Bytes[I ~int16 | ~uint16](i I) []byte {
	buf := make([]byte, 2)

	binary.BigEndian.PutUint16(buf, uint16(i))

	return buf
}

func I8Bytes[I ~int8 | ~uint8](i I) []byte {
	return []byte{byte(i)}
}

func F64Bytes[F ~float64](f F) []byte {
	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, math.Float64bits(float64(f)))

	return buf
}

func F32Bytes[F ~float32](f F) []byte {
	buf := make([]byte, 4)

	binary.BigEndian.PutUint32(buf, math.Float32bits(float32(f)))

	return buf
}

func SBytes[S ~string](s S) []byte {
	return []byte(s)
}
