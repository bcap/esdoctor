package util

import "strconv"

const kb float64 = 1024
const mb float64 = kb * 1024
const gb float64 = mb * 1024
const tb float64 = gb * 1024
const pb float64 = tb * 1024

func HumanizeBytes(numBytes int64) string {
	return HumanizeBytesF(float64(numBytes))
}

func HumanizeBytesF(numBytes float64) string {
	numBytesF := float64(numBytes)
	if numBytesF < kb {
		return strconv.FormatInt(int64(numBytesF), 10) + "b"
	} else if numBytesF < mb {
		return strconv.FormatFloat(numBytesF/kb, byte('f'), 1, 64) + "kb"
	} else if numBytesF < gb {
		return strconv.FormatFloat(numBytesF/mb, byte('f'), 1, 64) + "mb"
	} else if numBytesF < tb {
		return strconv.FormatFloat(numBytesF/gb, byte('f'), 1, 64) + "gb"
	} else if numBytesF < pb {
		return strconv.FormatFloat(numBytesF/tb, byte('f'), 1, 64) + "tb"
	} else {
		return strconv.FormatFloat(numBytesF/pb, byte('f'), 1, 64) + "pb"
	}
}
