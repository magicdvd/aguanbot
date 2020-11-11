package service

import (
	"errors"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ServiceManager 服务管理器
type ServiceManager struct {
	api API
	mgr *BotManager
}

var (
	smOnce                sync.Once
	instance              *ServiceManager
	errServiceDuplicated  = errors.New("regist duplicated service")
	errServiceUnsupported = errors.New("unsupported service")
)

func Mgr() *ServiceManager {
	smOnce.Do(func() {
		instance = &ServiceManager{}
	})
	return instance
}

func (c *ServiceManager) API() API {
	return c.api
}

func (c *ServiceManager) BotManager() *BotManager {
	return c.mgr
}

func (c *ServiceManager) Register(svcs ...interface{}) error {
	for _, svc := range svcs {
		switch s := svc.(type) {
		case API:
			if c.api != nil {
				return errServiceDuplicated
			}
			c.api = s
		case *BotManager:
			if c.mgr != nil {
				return errServiceDuplicated
			}
			c.mgr = s
		default:
			return errServiceUnsupported
		}
	}
	return nil
}
