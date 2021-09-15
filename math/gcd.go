package math

// overall algorithm:
//   https://www.khanacademy.org/computing/computer-science/cryptography/modarithmetic/a/the-euclidean-algorithm
// regarding negative inputs:
//   https://proofwiki.org/wiki/GCD_for_Negative_Integers
func GreatestCommonDivisor(a int64, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for {
		if a == 0 {
			return b
		}
		if b == 0 {
			return a
		}
		reminder := a % b
		a = b
		b = reminder
	}
}
