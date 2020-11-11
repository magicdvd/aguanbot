package service

import (
	"context"
	"testing"
)

func TestBot_Login(t *testing.T) {
	bot := newAguanAPI()
	a, err := bot.Login(context.Background(), "18600000000", "sd7155917")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(a)
	posts, a, err := bot.GetList(context.Background(), a)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(a)
	t.Log(posts)
}
