package service

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	CKDict    = "m:dict"
	CKAct     = "l:act"
	CKBotDict = "m:bots"
)

type Bot struct {
	Phone string
	*UserToken
	LastActivity time.Time
}

type userConfig struct {
	Phone string
	Pwd   string
}

func NewBot(phone string, password string) (*Bot, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	u, err := Mgr().API().Login(ctx, phone, password)
	if err != nil {
		return nil, err
	}
	return &Bot{
		Phone:        phone,
		UserToken:    u,
		LastActivity: time.Now(),
	}, nil
}

type BotManager struct {
	cache          *cache.Cache
	expireDuration time.Duration
	//log            *log.Factory
	botLock sync.RWMutex
}

func NewBotManager(users []string) *BotManager {
	dict := make(map[string]*userConfig)
	bots := make(map[string]*Bot)
	act := make([]string, 0)
	for _, v := range users {
		v = strings.Trim(v, "\n ")
		u := strings.Split(v, " ")
		if u[0] != "" && len(u) == 2 {
			dict[u[0]] = &userConfig{
				Phone: u[0],
				Pwd:   u[1],
			}
			act = append(act, u[0])
		}
	}
	cc := cache.New(10*time.Minute, 60*time.Minute)
	cc.Set(CKDict, dict, cache.NoExpiration)
	cc.Set(CKAct, act, cache.NoExpiration)
	cc.Set(CKBotDict, bots, cache.NoExpiration)
	return &BotManager{
		cache: cc,
	}
}

var errNoKey = errors.New("no cache key")
var errNoActUser = errors.New("no act users")

func (c *BotManager) Return(bot *Bot) {
	bots, ok := c.cache.Get(CKBotDict)
	if !ok {
		return
	}
	botsMap := bots.(map[string]*Bot)
	c.botLock.Lock()
	bot.LastActivity = time.Now()
	botsMap[bot.Phone] = bot
	c.botLock.Unlock()
	c.cache.Set(CKBotDict, botsMap, cache.NoExpiration)
}

func (c *BotManager) Get() (*Bot, error) {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	//idx := rd.Intn(c.actCount)
	act, ok := c.cache.Get(CKAct)
	if !ok {
		return nil, errNoKey
	}
	actUsers := act.([]string)
	l := len(actUsers)
	if l == 0 {
		return nil, errNoActUser
	}
	idx := rd.Intn(l)
	userPhone := actUsers[idx]
	actUsers = append(actUsers[:idx], actUsers[idx+1:]...)
	c.cache.Set(CKAct, actUsers, cache.NoExpiration)
	bots, ok := c.cache.Get(CKBotDict)
	if !ok {
		return nil, errNoKey
	}
	var retBot *Bot
	botsMap := bots.(map[string]*Bot)
	//缓存
	if u, ok := botsMap[userPhone]; ok {
		if time.Since(u.LastActivity) < c.expireDuration {
			retBot = u
		}
	}
	//存储缓存
	if retBot == nil {
		userConfigs, ok := c.cache.Get(CKDict)
		if !ok {
			return nil, errNoKey
		}
		cfg := userConfigs.(map[string]*userConfig)[userPhone]
		var err error
		retBot, err = NewBot(cfg.Phone, cfg.Pwd)
		if err != nil {
			return nil, err
		}
		c.botLock.Lock()
		botsMap[userPhone] = retBot
		c.botLock.Unlock()
		c.cache.Set(CKBotDict, botsMap, cache.NoExpiration)
	}
	return retBot, nil
}