package math

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercentilesBadInput(t *testing.T) {
	// bad input, valid buckets
	assert.Nil(t, PercentilesInt(nil, 10))
	assert.Nil(t, PercentilesInt([]int{}, 10))
	// valid input, bad buckets
	assert.Nil(t, PercentilesInt([]int{1, 2, 3, 4}, -1000))
	assert.Nil(t, PercentilesInt([]int{1, 2, 3, 4}, -1))
	assert.Nil(t, PercentilesInt([]int{1, 2, 3, 4}, 0))
	// bare minimum but still valid input:
	// an slice of only 1 value with 1 bucket should return a min=<value>, max=<value> result
	v := 100
	assert.Equal(t, []int{v, v}, PercentilesInt([]int{v}, 1))
}

func TestPercentiles(t *testing.T) {
	testInt := func(input []int, min int, p10 int, p50 int, p90 int, max int) {
		output := PercentilesInt(input, 10) // 10 points step
		assert.Equal(t, 10+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[1], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[5], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[9], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[10], "max of %v should be %v, got %v instead", output, max, output[10])

		output = PercentilesInt(input, 100) // 1 point step
		assert.Equal(t, 100+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[10], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[50], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[90], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[100], "max of %v should be %v, got %v instead", output, max, output[10])

		output = PercentilesInt(input, 1000) // 0.1 point step
		assert.Equal(t, 1000+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[100], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[500], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[900], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[1000], "max of %v should be %v, got %v instead", output, max, output[10])
	}

	testInt64 := func(input []int64, min int64, p10 int64, p50 int64, p90 int64, max int64) {
		output := PercentilesInt64(input, 10) // 10 points step
		assert.Equal(t, 10+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[1], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[5], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[9], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[10], "max of %v should be %v, got %v instead", output, max, output[10])

		output = PercentilesInt64(input, 100) // 1 point step
		assert.Equal(t, 100+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[10], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[50], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[90], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[100], "max of %v should be %v, got %v instead", output, max, output[10])

		output = PercentilesInt64(input, 1000) // 0.1 point step
		assert.Equal(t, 1000+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[100], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[500], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[900], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[1000], "max of %v should be %v, got %v instead", output, max, output[10])
	}

	testFloat64 := func(input []float64, min float64, p10 float64, p50 float64, p90 float64, max float64) {
		output := PercentilesFloat64(input, 10) // 10 points step
		assert.Equal(t, 10+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[1], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[5], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[9], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[10], "max of %v should be %v, got %v instead", output, max, output[10])

		output = PercentilesFloat64(input, 100) // 1 point step
		assert.Equal(t, 100+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[10], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[50], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[90], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[100], "max of %v should be %v, got %v instead", output, max, output[10])

		output = PercentilesFloat64(input, 1000) // 0.1 point step
		assert.Equal(t, 1000+1, len(output), "output has wrong length of %d", len(output))
		assert.Equal(t, min, output[0], "min of %v should be %v, got %v instead", output, min, output[0])
		assert.Equal(t, p10, output[100], "p10 of %v should be %v, got %v instead", output, p10, output[1])
		assert.Equal(t, p50, output[500], "p50 of %v should be %v, got %v instead", output, p50, output[5])
		assert.Equal(t, p90, output[900], "p90 of %v should be %v, got %v instead", output, p90, output[9])
		assert.Equal(t, max, output[1000], "max of %v should be %v, got %v instead", output, max, output[10])
	}

	testAllTypes := func(input []int, min int, p10 int, p50 int, p90 int, max int) {
		// fmt.Println(input)
		testInt(input, min, p10, p50, p90, max)

		input64 := make([]int64, len(input))
		for i, v := range input {
			input64[i] = int64(v)
		}
		testInt64(input64, int64(min), int64(p10), int64(p50), int64(p90), int64(max))

		inputF64 := make([]float64, len(input))
		for i, v := range input {
			inputF64[i] = float64(v)
		}
		testFloat64(inputF64, float64(min), float64(p10), float64(p50), float64(p90), float64(max))
	}

	test := func(input []int, min int, p10 int, p50 int, p90 int, max int) {
		testAllTypes(input, min, p10, p50, p90, max)

		// shuffling an array should not change its percentile results
		random := rand.New(rand.NewSource(1))
		for i := 0; i < 100; i++ {
			random.Shuffle(len(input), func(i, j int) { input[i], input[j] = input[j], input[i] })
			testAllTypes(input, min, p10, p50, p90, max)
		}

		// Duplicate the input array now with the shuffled elements.
		// Percentiles should not change if all elements in the array are duplicated
		tmp := make([]int, len(input)*2)
		copy(tmp[:len(input)], input)
		copy(tmp[len(input):], input)
		input = tmp
		testAllTypes(input, min, p10, p50, p90, max)
	}

	// small inputs
	test(
		[]int{0, 1},
		0, 0, 1, 1, 1,
	)
	test(
		[]int{0, 1, 2},
		0, 0, 1, 2, 2,
	)
	test(
		[]int{0, 1, 2, 3},
		0, 0, 2, 3, 3,
	)
	test(
		[]int{0, 1, 2, 3, 4},
		0, 0, 2, 4, 4,
	)

	// completely homogeneous input
	input := make([]int, 10000)
	test(input, 0, 0, 0, 0, 0)
	for idx, _ := range input {
		input[idx] = 520
	}
	test(input, 520, 520, 520, 520, 520)

	// homegeneous input but with small variations
	input[7923]++
	input[8020]--
	test(input, 520-1, 520, 520, 520, 520+1)

	// more meaningful inputs
	input = []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	test(input, 0, 10, 50, 90, 100)
}
