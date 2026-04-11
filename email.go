package main

import (
	"net/mail"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"golang.org/x/sync/errgroup"
)

func notifySuperusers(app core.App, title, msg string) (err error) {
	logger := app.Logger()
	defer err0.Then(&err, nil, func() {
		logger.Error("提醒邮件发送失败", "error", err)
	})
	users := try.To1(app.FindAllRecords(core.CollectionNameSuperusers, dbx.HashExp{"notify_muted": false}))
	if len(users) == 0 {
		return nil
	}
	eg := new(errgroup.Group)
	m := app.NewMailClient()
	am := app.Settings().Meta
	base := mailer.Message{
		From: mail.Address{
			Name:    am.SenderName,
			Address: am.SenderAddress,
		},
		Subject: title,
		Text:    msg,
	}
	for _, user := range users {
		msg := base
		msg.To = []mail.Address{
			{Address: user.Email()},
		}
		eg.Go(func() error {
			return m.Send(&msg)
		})
	}
	return eg.Wait()
}
