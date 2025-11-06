package aliyun

const PageSize = 500

func loop[T any](size int, fn func(i int) ([]T, error)) (items []T, err error) {
	for i := 1; true; i++ {
		var n []T
		n, err = fn(i)
		if err != nil {
			return
		}
		items = append(items, n...)
		if len(n) < size {
			break
		}
	}
	return
}
