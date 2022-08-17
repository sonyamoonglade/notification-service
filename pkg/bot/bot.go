package bot

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Config struct {
	BotToken string
}

type Bot interface {
	Notify(receiverID int64, fmtTempl string) error
	GetClient() *tg.BotAPI
	GetUpdatesCfg() tg.UpdateConfig
}

type bot struct {
	client    *tg.BotAPI
	logger    *zap.SugaredLogger
	updateCfg tg.UpdateConfig
}

func NewBot(cfg Config, logger *zap.SugaredLogger) (Bot, error) {

	client, err := tg.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	client.Debug = true

	updateCfg := tg.NewUpdate(0)
	updateCfg.Timeout = 60

	return &bot{
		logger:    logger,
		client:    client,
		updateCfg: updateCfg,
	}, nil
}

func (b *bot) Notify(receiverID int64, fmtTempl string) error {
	//TODO implement me
	panic("implement me")
}

func (b *bot) GetClient() *tg.BotAPI {
	return b.client
}

func (b *bot) GetUpdatesCfg() tg.UpdateConfig {
	return b.updateCfg

}
