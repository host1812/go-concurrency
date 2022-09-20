package main

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"testing"
)

func Test_some(t *testing.T) {
	var str bytes.Buffer
	log.SetOutput(&str)
	var wg sync.WaitGroup
	wg.Add(1)
	go some("test1", &wg)
	wg.Wait()
	expected := "some: finished"
	if !strings.Contains(str.String(), expected) {
		t.Errorf("expected: %s\ngot: %s", expected, str.String())
	}
}
