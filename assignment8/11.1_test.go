package main

import ( 
	"bytes"
  "testing"
)

func TestGenerate1(t *testing.T ) {
	main()

	var stdin bytes.Buffer

  stdin.Write([]byte("hunter2\n"))

  result, err := cmd.Run(&stdin)
  assert.NoError(t, err)
  assert.Equal(t, "hunter2", result)
}