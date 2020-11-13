package service

import (
	"context"
	"testing"

	"github.com/whatisfaker/zaptrace/log"
)

func TestBot_Login(t *testing.T) {
	a := NewAguanAPI(log.NewStdLogger("debug"))
	b, err := a.Login(context.TODO(), "18600000000", "sd7155917")
	if err != nil {
		t.Error(err)
	}
	t.Log(b)
}
