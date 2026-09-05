package validations

func CalculateTotalPages(total int64, pageSize int) int64 {
	if total == 0 {
		return 0
	}

	return (total + int64(pageSize) - 1) / int64(pageSize)
}
