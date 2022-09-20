package main

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func Test_main(t *testing.T) {
	b := bytes.Buffer{}
	log.SetOutput(&b)
	expected := "balance final: 254997476520"
	main()
	if !strings.Contains(b.String(), expected) {
		t.Errorf("expected: %s, got: %s", expected, b.String())
	}

}
