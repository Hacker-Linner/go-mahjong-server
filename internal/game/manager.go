package game

import (
	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
)

var defaultManager = NewManager()

type (
	Manager struct {
		component.Base
		group *nano.Group // 广播channel
	}
)

func NewManager() *Manager {
	return &Manager{
		group: nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
	}
}

func (c *Manager) DemoHandler(s *session.Session, raw []byte) error {
	// 业务逻辑开始
	// ...
	// 业务逻辑结束
	return nil
}
