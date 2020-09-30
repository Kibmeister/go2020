package main

import ( "testing")

func test_dial(t *testing.T) {
	_, addr := createListener()
	go dial(addr)
}
