package osc_test

import (
	"fmt"
	"testing"

	osc "github.com/vchrisr/go-osc"
)

func TestAutoDiscover(t *testing.T) {
	discoveries, err := osc.AutoDiscover(10023, osc.NewMessage("/info"))
	if err != nil {
		t.Errorf("Expected err to be nil but got %v", err)
	}

	for i, d := range discoveries {
		fmt.Println(i, d)
	}
}
