package nepse

var dataSegment = []int32{
	0x05, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x07, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x06, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
	0x05, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
	0x03, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00,
	0x06, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
	0x06, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00,
	0x05, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x09, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00,
	0x08, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x09,
}

func cdx(w2cP0, w2cP1 int32) int32 {
	w2cI0 := w2cP1
	w2cI1 := int32(10)
	w2cI0 = w2cI0 / w2cI1
	w2cP0 = w2cI0
	w2cI0 = w2cI0 % w2cI1
	w2cI1 = w2cP1
	w2cI2 := w2cP0
	w2cI3 := int32(10)
	w2cI2 *= w2cI3
	w2cI1 -= w2cI2
	w2cI0 += w2cI1
	w2cI1 = w2cP1
	w2cI2 = int32(100)
	w2cI1 = w2cI1 / w2cI2
	w2cI2 = int32(10)
	w2cI1 = w2cI1 % w2cI2
	w2cI0 += w2cI1
	w2cI1 = int32(2)
	w2cI0 <<= w2cI1 & 31

	w2cI0 = dataSegment[w2cI0]
	w2cI1 = int32(22)
	w2cI0 += w2cI1

	return w2cI0
}

func rdx(w2cP0, w2cP1, w2cP2 int32) int32 {
	w2cI0 := w2cP1
	w2cI1 := int32(100)
	w2cI0 = w2cI0 / w2cI1
	w2cI1 = int32(10)
	w2cI0 = w2cI0 % w2cI1
	w2cI1 = w2cP1
	w2cI2 := int32(10)
	w2cI1 = w2cI1 / w2cI2
	w2cP0 = w2cI1
	w2cI2 = int32(10)
	w2cI1 = w2cI1 % w2cI2
	w2cI0 += w2cI1
	w2cP2 = w2cI0
	w2cI1 = w2cP2
	w2cI2 = w2cP1
	w2cI3 := w2cP0
	w2cI4 := int32(10)
	w2cI3 *= w2cI4
	w2cI2 -= w2cI3
	w2cI1 += w2cI2
	w2cI2 = int32(2)
	w2cI1 <<= w2cI2 & 31

	w2cI1 = dataSegment[w2cI1]
	w2cI0 += w2cI1
	w2cI1 = int32(32)
	w2cI0 += w2cI1

	return w2cI0
}

func bdx(w2cP0, w2cP1, w2cP2 int32) int32 {
	w2cI0 := w2cP1
	w2cI1 := int32(100)
	w2cI0 = w2cI0 / w2cI1
	w2cI1 = int32(10)
	w2cI0 = w2cI0 % w2cI1
	w2cI1 = w2cP1
	w2cI2 := int32(10)
	w2cI1 = w2cI1 / w2cI2
	w2cP0 = w2cI1
	w2cI2 = int32(10)
	w2cI1 = w2cI1 % w2cI2
	w2cI0 += w2cI1
	w2cP2 = w2cI0
	w2cI1 = w2cP2
	w2cI2 = w2cP1
	w2cI3 := w2cP0
	w2cI4 := int32(10)
	w2cI3 *= w2cI4
	w2cI2 -= w2cI3
	w2cI1 += w2cI2
	w2cI2 = int32(2)
	w2cI1 <<= w2cI2 & 31

	w2cI1 = dataSegment[w2cI1]
	w2cI0 += w2cI1
	w2cI1 = int32(60)
	w2cI0 += w2cI1

	return w2cI0
}

func ndx(w2cP0, w2cP1, w2cP2 int32) int32 {
	w2cI0 := w2cP1
	w2cI1 := int32(10)
	w2cI0 = w2cI0 / w2cI1
	w2cP0 = w2cI0
	w2cI1 = int32(10)
	w2cI0 = w2cI0 % w2cI1
	w2cP2 = w2cI0
	w2cI1 = w2cP2
	w2cI2 := w2cP1
	w2cI3 := w2cP0
	w2cI4 := int32(10)
	w2cI3 *= w2cI4
	w2cI2 -= w2cI3
	w2cI1 += w2cI2
	w2cI2 = w2cP1
	w2cI3 = int32(100)
	w2cI2 = w2cI2 / w2cI3
	w2cI3 = int32(10)
	w2cI2 = w2cI2 % w2cI3
	w2cI1 += w2cI2
	w2cI2 = int32(2)
	w2cI1 <<= w2cI2 & 31

	w2cI1 = dataSegment[w2cI1]
	w2cI0 += w2cI1
	w2cI1 = int32(88)
	w2cI0 += w2cI1

	return w2cI0
}

func mdx(w2cP0, w2cP1, w2cP2 int32) int32 {
	w2cI0 := w2cP1
	w2cI1 := int32(100)
	w2cI0 = w2cI0 / w2cI1
	w2cI1 = int32(10)
	w2cI0 = w2cI0 % w2cI1
	w2cP0 = w2cI0
	w2cI1 = w2cP0
	w2cI2 := w2cP1
	w2cI3 := int32(10)
	w2cI2 = w2cI2 / w2cI3
	w2cP2 = w2cI2
	w2cI3 = int32(10)
	w2cI2 = w2cI2 % w2cI3
	w2cI3 = w2cP1
	w2cI4 := w2cP2
	w2cI5 := int32(10)
	w2cI4 *= w2cI5
	w2cI3 -= w2cI4
	w2cI2 += w2cI3
	w2cI1 += w2cI2
	w2cI2 = int32(2)
	w2cI1 <<= w2cI2 & 31

	w2cI1 = dataSegment[w2cI1]
	w2cI0 += w2cI1
	w2cI1 = int32(110)
	w2cI0 += w2cI1

	return w2cI0
}

// 202412021365
// A5uT2X7n
