package event

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/magicdvd/aguanbot/service"
	"github.com/whatisfaker/zaptrace/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Activity struct {
	rd   *rand.Rand
	pMax int           //最多几个帖子
	dur  time.Duration //一个事件的持续时间
	log  *log.Factory
}

func NewActivity(postMax int, dur time.Duration, log *log.Factory) *Activity {
	return &Activity{
		pMax: postMax,
		dur:  dur,
		log:  log,
	}
}

func (c *Activity) doVote(ctx context.Context, post *service.Post) error {
	//c.log.Normal().Debug("do vote start", zap.Uint("post_id", post.ID))
	bot, err := service.Mgr().BotManager().Get()
	if err != nil {
		c.log.Normal().Error("doVote(GetBot)", zap.Error(err))
		return err
	}
	l := len(post.Tags)
	if l == 0 {
		c.log.Normal().Warn("doVote(Exit)", zap.String("reason", "no tags"), zap.Uint("post_id", post.ID))
		return nil
	}
	idx := c.rd.Intn(l)
	tag := post.Tags[idx]
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	tk, err := service.Mgr().API().AgreeTag(ctx, bot.UserToken, tag.ID)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		c.log.Normal().Error("doVote(AgreeTag)", zap.Error(err))
		return err
	}
	if err == nil {
		bot.UserToken = tk
		service.Mgr().BotManager().Return(bot)
	}
	c.log.Normal().Debug("do vote complete", zap.Uint("post_id", post.ID), zap.Uint("tag_id", tag.ID), zap.String("user", bot.Phone))
	return nil
}

func (c *Activity) Do(pctx context.Context) error {
	c.log.Trace(pctx).Debug("activity start")
	bot, err := service.Mgr().BotManager().Get()
	if err != nil {
		c.log.Normal().Error("Do(GetBot)", zap.Error(err))
		return err
	}
	posts, tk, err := service.Mgr().API().GetList(pctx, bot.UserToken)
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
	grp, ctx := errgroup.WithContext(pctx)
	for _, v := range doPosts {
		post := v
		grp.Go(func() error {
			c.log.Normal().Debug("vote first", zap.Uint("post_id", post.ID))
			err := c.doVote(ctx, &post)
			if err != nil {
				c.log.Normal().Error("Do(doVote)", zap.Error(err))
				return err
			}
			userCount := service.Mgr().BotManager().CalcBotForPost(v.TagAgreeCount, tcc)
			sendCount := make([]int, 0)
			tmp := userCount
			for tmp > 0 {
				ic := 1
				if tmp > 1 {
					ic = c.rd.Intn(tmp-1) + 1
				}
				sendCount = append(sendCount, ic)
				tmp -= ic
			}
			waitInterval := c.dur / time.Duration(userCount)
			c.log.Normal().Debug("vote next", zap.Uint("post_id", post.ID), zap.Int("user_count", userCount), zap.Duration("average_wait", waitInterval))
			start := time.Now()
			for i, voteCount := range sendCount {
				for voteCount > 0 {
					voteCount--
					t := waitInterval
					wait := c.rd.Int63n(int64(t))
					wait1 := time.Until(start.Add(time.Duration(i)*t + time.Duration(wait)))
					c.log.Normal().Debug("vote next wait", zap.Uint("post_id", post.ID), zap.Int("user_index", i), zap.Duration("wait", time.Duration(wait1)))
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(wait1):
					}
					c.log.Normal().Debug("vote follow", zap.Uint("post_id", post.ID))
					err := c.doVote(ctx, &post)
					if err != nil {
						c.log.Normal().Error("Do(doVote)", zap.Error(err))
						return err
					}
					wait2 := time.Until(start.Add(time.Duration(i+1) * t))
					c.log.Normal().Debug("vote next wait2", zap.Uint("post_id", post.ID), zap.Int("user_index", i), zap.Duration("wait", time.Duration(wait2)))
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(wait2):
					}
				}
			}
			return nil
		})
		wait := c.dur / time.Duration(c.pMax)
		c.log.Normal().Debug("vote post wait", zap.Duration("wait", time.Duration(wait)))
		select {
		case <-pctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return grp.Wait()
}
