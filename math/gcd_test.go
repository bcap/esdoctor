package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGreatestCommonDivisor(t *testing.T) {
	test := func(a int64, b int64, expected int64) {
		testSingle := func(a int64, b int64) {
			got := GreatestCommonDivisor(a, b)
			assert.Equal(t, expected, got, "GCD(%d, %d) should be %d but got %d instead", a, b, expected, got)
		}

		for _, f := range []int64{1, -1} {
			testSingle(a*f, b)
			testSingle(a*f, b*f)
			testSingle(a, b*f)

			testSingle(b*f, a)
			testSingle(b*f, a*f)
			testSingle(b, a*f)
		}
	}

	// > However, zero is its own greatest divisor if greatest is understood in
	// > the context of the divisibility relation, so gcd(0, 0) is commonly defined as 0.
	// from https://en.wikipedia.org/wiki/Greatest_common_divisor
	test(0, 0, 0)

	test(1, 1, 1)
	test(2, 2, 2)
	test(2, 4, 2)
	test(10, 25, 5)
	test(270, 192, 6)

	var large int64 = 250723525
	test(large, large*(large+1), large)
}
