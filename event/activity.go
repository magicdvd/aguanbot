package event

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/magicdvd/aguanbot/service"
	"github.com/whatisfaker/zaptrace/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	smOnce   sync.Once
	instance *Activity
)

func Event() *Activity {
	smOnce.Do(func() {
		instance = &Activity{}
	})
	return instance
}

type Activity struct {
	rd   *rand.Rand
	pMax int           //最多几个帖子
	dur  time.Duration //一个事件的持续时间
	log  *log.Factory
}

func (c *Activity) doVote(ctx context.Context, post *service.Post) error {
	c.log.Trace(ctx).Debug("vote post", zap.Uint("post_id", post.ID))
	bot, err := service.Mgr().BotManager().Get()
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
	tk, err := service.Mgr().API().AgreeTag(ctx, bot.UserToken, tag.ID)
	if err != nil {
		c.log.Normal().Error("Do(AgreeTag)", zap.Error(err))
		return err
	}
	bot.UserToken = tk
	service.Mgr().BotManager().Return(bot)
	return nil
}

func (c *Activity) Do(ctx context.Context) error {
	c.log.Trace(ctx).Debug("activity start")
	bot, err := service.Mgr().BotManager().Get()
	if err != nil {
		c.log.Normal().Error("Do(GetBot)", zap.Error(err))
		return err
	}
	posts, tk, err := service.Mgr().API().GetList(ctx, bot.UserToken)
	if err != nil {
		c.log.Normal().Error("Do(GetList)", zap.Error(err))
		return err
	}
	bot.UserToken = tk
	service.Mgr().BotManager().Return(bot)
	l := len(posts)
	if l == 0 {
		c.log.Normal().Warn("Do(Exit)", zap.String("reason", "no posts"))
		return nil
	}
	c.rd = rand.New(rand.NewSource(time.Now().UnixNano()))
	max := c.pMax
	if max > l {
		max = l
	}
	doPosts := make([]service.Post, 0)
	for max > 0 {
		idx := c.rd.Intn(l)
		post := posts[idx]
		if len(post.Tags) == 0 {
			continue
		}
		doPosts = append(doPosts, post)
		max--
	}
	var tcc int
	for i, v := range doPosts {
		v.TagAgreeCount = 0
		for _, t := range v.Tags {
			v.TagAgreeCount += t.Agreed
		}
		tcc += v.TagAgreeCount
		doPosts[i] = v
	}
	grp, ctx := errgroup.WithContext(ctx)
	for _, v := range doPosts {
		post := v
		grp.Go(func() error {
			err := c.doVote(ctx, &post)
			if err != nil {
				c.log.Normal().Error("Do(doVote)", zap.Error(err))
				return err
			}
			userCount := service.Mgr().BotManager().CalcBotForPost(v.TagAgreeCount, tcc)
			sendCount := make([]int, 0)
			for userCount > 0 {
				ic := 1
				if userCount > 1 {
					ic = c.rd.Intn(userCount-1) + 1
				}
				sendCount = append(sendCount, ic)
				userCount -= ic
			}
			start := time.Now()
			waitInterval := c.dur / time.Duration(userCount)
			for i, voteCount := range sendCount {
				for voteCount > 0 {
					voteCount--
					t := waitInterval
					wait := c.rd.Int63n(int64(t))
					time.Sleep(time.Duration(wait))
					err := c.doVote(ctx, &post)
					if err != nil {
						c.log.Normal().Error("Do(doVote)", zap.Error(err))
						return err
					}
					wait2 := time.Until(start.Add(time.Duration(i+1) * t))
					time.Sleep(wait2)
				}
			}
			return nil
		})
	}
	return grp.Wait()
}
