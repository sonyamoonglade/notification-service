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
	StartKeyboard() tg.ReplyKeyboardMarkup
	Send(ch tg.Chattable) (*tg.Message, error)
	SoftSend(ch tg.Chattable) error
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

	updateCfg := tg.NewUpdate(0)
	updateCfg.Timeout = 60

	return &bot{
		logger:    logger,
		client:    client,
		updateCfg: updateCfg,
	}, nil
}

//SoftSend it syntax-sugar for sends that do not require message to be returned
func (b *bot) SoftSend(ch tg.Chattable) error {
	_, err := b.Send(ch)
	return err
}

func (b *bot) Notify(receiverID int64, fmtTempl string) error {
	msg := tg.NewMessage(receiverID, fmtTempl)

	_, err := b.Send(msg)
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
func (b *bot) Send(ch tg.Chattable) (*tg.Message, error) {
	m, err := b.client.Send(ch)
	if err != nil {
		b.logger.Error(err.Error())
		//Todo: notify admin here
		return nil, fmt.Errorf("bot could not send a message. %s", err.Error())
	}
	return &m, nil
}

func (b *bot) StartKeyboard() tg.ReplyKeyboardMarkup {
	bt := tg.KeyboardButton{
		Text:           "Получать уведомления",
		RequestContact: true,
	}
	row := []tg.KeyboardButton{bt}
	return tg.NewReplyKeyboard(row)
}
