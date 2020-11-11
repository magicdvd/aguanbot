package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/whatisfaker/zaptrace/log"
	"go.uber.org/zap"
)

type Group struct {
	// duration time.Duration //持续时间
	// min      int
	rd   *rand.Rand
	pMax int //最多几个帖子
	uMax int //最多多少人
	log  *log.Factory
}

func (c *Group) doVote(ctx context.Context, post *Post) error {
	bot, err := Mgr().BotManager().Get()
	if err != nil {
		c.log.Normal().Error("Do(GetBot)", zap.Error(err))
		return err
	}
	l := len(post.Tags)
	if l == 0 {
		c.log.Normal().Warn("Do(Exit)", zap.String("reason", "no tags"), zap.Uint("post_id", post.ID))
		return nil
	}
	idx := c.rd.Intn(l)
	tag := post.Tags[idx]
	tk, err := Mgr().API().AgreeTag(ctx, bot.UserToken, tag.ID)
	if err != nil {
		c.log.Normal().Error("Do(AgreeTag)", zap.Error(err))
		return err
	}
	bot.UserToken = tk
	Mgr().BotManager().Return(bot)
	return nil
}

func (c *Group) Do(ctx context.Context) error {
	bot, err := Mgr().BotManager().Get()
	if err != nil {
		c.log.Normal().Error("Do(GetBot)", zap.Error(err))
		return err
	}
	posts, tk, err := Mgr().API().GetList(ctx, bot.UserToken)
	if err != nil {
		c.log.Normal().Error("Do(GetList)", zap.Error(err))
		return err
	}
	bot.UserToken = tk
	Mgr().BotManager().Return(bot)
	l := len(posts)
	if l == 0 {
		c.log.Normal().Warn("Do(Exit)", zap.String("reason", "no posts"))
		return nil
	}
	c.rd = rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := c.rd.Intn(l)
	post := posts[idx]
	err = c.doVote(ctx, &post)
	if err != nil {
		return err
	}
	return nil
}
