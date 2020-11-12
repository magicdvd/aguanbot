package service

import (
	"math/rand"
	"testing"
)

func TestBot(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6}
	a = append(a[0:2], a[3:]...)
	t.Log(a)
}

func TestBot1(t *testing.T) {
	ic := rand.Intn(0) + 1
	t.Log(ic)
}
