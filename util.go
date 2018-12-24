package rtspgo

import "github.com/pkg/errors"

// FindStrFromSlice []string中string的比对
func FindStrFromSlice(s string, ss []string) (int, error) {
	var index = -1

	if len(ss) == 0 {
		return -1, errors.Errorf("slice len is 0")
	}

	for i, elt := range ss {
		if s == elt {
			index = i
			break
		}
	}

	return index, nil
}
