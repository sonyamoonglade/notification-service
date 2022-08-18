package bot

import (
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

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

func NewBot(token string, logger *zap.SugaredLogger) (Bot, error) {

	client, err := tg.NewBotAPI(token)
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
	msg := tg.NewMessage(receiverID, fmtTempl)

	_, err := b.send(msg)
	if err != nil {
		return err
	}

	b.logger.Debugf("notified %d successfully", receiverID)
	return nil
}

func (b *bot) GetClient() *tg.BotAPI {
	return b.client
}

func (b *bot) GetUpdatesCfg() tg.UpdateConfig {
	return b.updateCfg

}
func (b *bot) send(ch tg.Chattable) (*tg.Message, error) {
	m, err := b.client.Send(ch)
	if err != nil {
		return nil, fmt.Errorf("bot could not send a message. %s", err.Error())
	}
	return &m, nil
}
