package storageerrs

import "fmt"

var (
	ErrNoResults   = fmt.Errorf("no rows found to match the request")
	ErrOrderExists = fmt.Errorf("order already belongs other customer")
	ErrOrderLoaded = fmt.Errorf("order already exists")
)
