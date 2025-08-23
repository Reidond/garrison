package utils

import "fmt"

func RequireNonEmpty(name, v string) error {
	if v == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	return nil
}

func RequirePositive(name string, v int) error {
	if v <= 0 {
		return fmt.Errorf("%s must be > 0", name)
	}
	return nil
}
