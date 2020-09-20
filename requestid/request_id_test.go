package requestid

import (
	"fmt"
	"testing"
)

func TestGenRequestID(t *testing.T) {
	fmt.Println(GenRequestID())
}

func TestParseRequestID(t *testing.T) {
	ParseRequestID(GenRequestID())
}
