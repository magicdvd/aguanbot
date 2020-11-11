package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/whatisfaker/zaptrace/log"
	"go.uber.org/zap"
)

type Group struct {
	duration time.Duration //持续时间
	min      int
	post     *Post
	log      *log.Factory
}

func (c *Group) doVote(ctx context.Context) error {
	bot, err := Mgr().BotManager().Get()
	if err != nil {
		c.log.Normal().Error("Do(GetBot)", zap.Error(err))
		return err
	}
	l := len(c.post.Tags)
	if l == 0 {
		c.log.Normal().Warn("Do(Exit)", zap.String("reason", "no tags"), zap.Uint("post_id", c.post.ID))
		return nil
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rd.Intn(l)
	tag := c.post.Tags[idx]
	tk, err := Mgr().API().AgreeTag(ctx, bot.UserToken, tag.ID)
	if err != nil {
		c.log.Normal().Error("Do(AgreeTag)", zap.Error(err))
		return err
	}
	bot.UserToken = tk
	Mgr().BotManager().Return(bot)
	return nil
}
