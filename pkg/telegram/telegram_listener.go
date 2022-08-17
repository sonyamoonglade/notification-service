package telegram

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
)

type Listener interface {
	ListenForUpdates(bot *bot.Bot)
	handleContact(cnt *tg.Contact)
	handleMessage(msg *tg.Message)
	mapUpdate(upd *tg.Update)
}

type telegramListener struct {
}

func (t *telegramListener) HandleContact(cnt *tg.Contact) {
	//TODO implement me
	panic("implement me")
}

func (t *telegramListener) HandleMessage(msg *tg.Message) {
	//TODO implement me
	panic("implement me")
}

func (t *telegramListener) Map(upd *tg.Update) {
	//TODO implement me
	panic("implement me")
}

func (t *telegramListener) ListenForUpdates(bot bot.Bot) {
	cl := bot.GetClient()
	cfg := bot.GetUpdatesCfg()
	updch := cl.GetUpdatesChan(cfg)
	for upd := range updch {
		t.Map(&upd)
	}
}
