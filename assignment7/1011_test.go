package main

import ( 
	"testing" 
)

func TestGenerate1(t *testing.T ) {
	Generate("filename", "password")
	Sign("filename", "password", []byte("hello you"))
}

func TestGenerate2(t *testing.T ) {
	Generate("Filename", "password")
	Sign("filename", "password", []byte("hello you"))
}

func TestGenerate3(t *testing.T ) {
	Generate("filenamefilename", "p")
	Sign("filenamefilename", "p", []byte("hello you"))
}

func TestGenerate4(t *testing.T ) {
	Generate("filenfilename", "p")
	Sign("filenamefilename", "p", []byte("hello you"))
}

func TestGenerate5(t *testing.T ) {
	Generate("filename", "pasas")
	Sign("filename", "p", []byte("hello you"))
}

func TestGenerate6(t *testing.T ) {
	Generate("", "pasas")
	Sign("filename", "p", []byte("hello you"))
}

func TestGenerate7(t *testing.T ) {
	Generate("123", "pasas")
	Sign("filename", "p", []byte("hello you"))
}

