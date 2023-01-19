package orders

import (
	"strconv"
	"time"
)

type (
	Status struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}

	Withdrawal struct {
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at,omitempty"`
	}
)

func OrderNumValid(orderNum string) bool {
	origNumber, err := strconv.ParseInt(orderNum, 10, 64)
	if err != nil {
		return false
	}

	// Luhn algorithm validation
	var luhn int64

	control := origNumber % 10
	number := origNumber / 10

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number /= 10
	}

	return (control+luhn%10)%10 == 0
}
