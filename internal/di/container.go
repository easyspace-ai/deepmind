package di

import (
	"github.com/weibaohui/nanobot-go/internal/infrastructure/eventbus"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	Logger   *zap.Logger
	DB       *gorm.DB
	EventBus *eventbus.SimpleEventBus
}

// NewContainer 创建新的 DI 容器
func NewContainer(db *gorm.DB, logger *zap.Logger) *Container {
	return &Container{
		Logger:   logger,
		DB:       db,
		EventBus: eventbus.NewSimpleEventBus(),
	}
}
