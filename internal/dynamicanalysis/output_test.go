package dynamicanalysis

import (
	"bytes"
	"testing"
)

func TestEmptyBuffer(t *testing.T) {
	b := []byte{}
	if test := lastLines(b, 10, 100); len(test) > 0 {
		t.Fatalf("lastLines() = %v; want empty byte array", test)
	}
}

func TestSmallBuffer(t *testing.T) {
	b := []byte("hello world")
	if test := lastLines(b, 10, 100); !bytes.Equal(test, b) {
		t.Fatalf("lastLines() = %v, want %v", test, b)
	}
}

func TestSmallBufferWithFewLines(t *testing.T) {
	b := []byte("hello\nworld\n")
	if test := lastLines(b, 10, 100); !bytes.Equal(test, b) {
		t.Fatalf("lastLines() = %v, want %v", test, b)
	}
}

func TestSmallBufferWithManyLines(t *testing.T) {
	b := []byte("one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
	exp := []byte("six\nseven\neight\nnine\nten\n")
	if test := lastLines(b, 5, 100); !bytes.Equal(test, exp) {
		t.Fatalf("lastLines() = %v, want %v", test, exp)
	}
}

func TestBigBufferWithNoLines(t *testing.T) {
	b := make([]byte, 1000, 1000)
	for i := range b {
		b[i] = 'a' + byte(i%26)
	}
	exp := []byte("qrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl")
	if test := lastLines(b, 10, 100); !bytes.Equal(test, exp) {
		t.Fatalf("lastLines() = %v (%d), want %v (%d)", test, len(test), exp, len(exp))
	}
}
