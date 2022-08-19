package telegram

import (
	"context"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/notification-service/internal/subscription"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"github.com/sonyamoonglade/notification-service/pkg/message"
	"github.com/sonyamoonglade/notification-service/pkg/tgErrors"
	"go.uber.org/zap"
)

type Listener interface {
	ListenForUpdates()
	handleContact(ctx context.Context, chatID int64, cnt *tg.Contact)
	handleMessage(ctx context.Context, chatID int64, msg *tg.Message)
	mapUpdate(upd *tg.Update)
}

type telegramListener struct {
	logger              *zap.SugaredLogger
	bot                 bot.Bot
	subscriptionService subscription.Service
}

func NewTelegramListener(logger *zap.SugaredLogger, bot bot.Bot, subscriptionService subscription.Service) Listener {
	return &telegramListener{logger: logger, bot: bot, subscriptionService: subscriptionService}
}

func (t *telegramListener) handleContact(ctx context.Context, chatID int64, cnt *tg.Contact) {

	phoneNumber := cnt.PhoneNumber

	//Get subscriber
	sub, err := t.subscriptionService.GetSubscriberByPhone(ctx, phoneNumber)
	if err != nil {
		//Exit here in case there's no registered subscriber by given number
		if errors.Is(err, httpErrors.ErrSubscriberDoesNotExist) {
			text := message.Format(message.NoSuchSubscriber, phoneNumber)
			msg := tg.NewMessage(chatID, text)

			err := t.bot.SoftSend(msg)
			if err != nil {
				return
			}
		}
		//Some internal error
		t.logger.Error(err.Error())
		msg := tg.NewMessage(chatID, message.SomethingWentWrong)
		err := t.bot.SoftSend(msg)
		if err != nil {
			return
		}

		return
	}

	//Show registering process
	msg1 := tg.NewMessage(chatID, message.RegisterInProcess)
	err = t.bot.SoftSend(msg1)
	if err != nil {
		return
	}

	//Try register telegramSubscriber (might fail with ErrTelegramSubscriberExists error)
	//In this step we assign chatID of the user with phoneNumber of subscriber
	err = t.subscriptionService.RegisterTelegramSubscriber(ctx, chatID, sub.SubscriberID)
	if err != nil {
		//Bot already knows specified telegramSubscriber by given phoneNumber
		if errors.Is(err, tgErrors.ErrTgSubscriberAlreadyExists) {
			text := message.Format(message.IKnowYou, phoneNumber)
			msg := tg.NewMessage(chatID, text)

			err = t.bot.SoftSend(msg)
			if err != nil {
				return
			}
			//End execution
			return
		}
		t.logger.Error(err.Error())
		//Something went wrong internally
		msg := tg.NewMessage(chatID, message.SomethingWentWrong)

		err = t.bot.SoftSend(msg)
		if err != nil {
			return
		}
		return
	}

	//Successfully registered subscriber
	text2 := message.Format(message.RegisteredTelegramSubscriber, phoneNumber)
	msg2 := tg.NewMessage(chatID, text2)

	err = t.bot.SoftSend(msg2)
	if err != nil {
		return
	}
}

func (t *telegramListener) handleMessage(ctx context.Context, chatID int64, m *tg.Message) {

	switch m.Text {
	case "/start":
		startKb := t.bot.StartKeyboard()
		msg := tg.NewMessage(chatID, message.StartMessage)
		msg.ReplyMarkup = startKb

		startKb.InputFieldPlaceholder = "Получать уведомления"
		startKb.OneTimeKeyboard = false
		startKb.ResizeKeyboard = false

		err := t.bot.SoftSend(msg)
		if err != nil {
			return
		}

		return
	default:
		//Ignore other messages...
		return
	}

}

func (t *telegramListener) mapUpdate(upd *tg.Update) {
	isGroup := upd.FromChat().IsGroup()
	if isGroup {
		//Ignore group messages at all
		return
	}
	chatID := upd.FromChat().ID

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	switch true {
	case upd.Message != nil && upd.Message.Contact != nil:
		t.handleContact(ctx, chatID, upd.Message.Contact)
		return
	case upd.Message != nil:
		t.handleMessage(ctx, chatID, upd.Message)
		return
	}
}

func (t *telegramListener) ListenForUpdates() {
	cl := t.bot.GetClient()

	cfg := t.bot.GetUpdatesCfg()
	updch := cl.GetUpdatesChan(cfg)
	for upd := range updch {
		t.mapUpdate(&upd)
	}
}
