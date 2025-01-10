package slicex

func Map[A, B any](input []A, f func(A) B) []B {
	count := len(input)
	res := make([]B, count)
	for i := 0; i < count; i++ {
		res[i] = f(input[i])
	}
	return res
}
