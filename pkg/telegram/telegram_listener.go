package telegram

import (
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
	"go.uber.org/zap"
)

type Listener interface {
	ListenForUpdates()
	handleContact(cnt *tg.Contact)
	handleMessage(msg *tg.Message)
	mapUpdate(upd *tg.Update)
}

type telegramListener struct {
	logger *zap.SugaredLogger
	bot    bot.Bot
}

func NewTelegramListener(logger *zap.SugaredLogger, bot bot.Bot) Listener {
	return &telegramListener{logger: logger, bot: bot}
}

func (t *telegramListener) handleContact(cnt *tg.Contact) {
	//TODO implement me
	panic("implement me")
}

func (t *telegramListener) handleMessage(msg *tg.Message) {
	//TODO implement me
	panic("implement me")
}

func (t *telegramListener) mapUpdate(upd *tg.Update) {
	fmt.Println("map update")
}

func (t *telegramListener) ListenForUpdates() {
	cl := t.bot.GetClient()
	cfg := t.bot.GetUpdatesCfg()
	updch := cl.GetUpdatesChan(cfg)
	for upd := range updch {
		t.mapUpdate(&upd)
	}
}
