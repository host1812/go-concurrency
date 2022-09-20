package main

import "testing"

func Test_update(t *testing.T) {
	msg = "test"
	wg.Add(1)
	go update2("a", &wg)
	wg.Wait()

	if msg != "a" {
		t.Errorf("not passed")
	}
}
