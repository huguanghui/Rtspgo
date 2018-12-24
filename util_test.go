package rtspgo

import (
	"fmt"
	"testing"
)

func TestFindStrFromSlice(t *testing.T) {
	e := "abe"
	es := []string{"abc", "abd", "abe"}
	// es := []string{}

	i, err := FindStrFromSlice(e, es)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("find index: ", i)
}
