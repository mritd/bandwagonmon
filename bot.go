package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
	"strings"
	"time"
)

func (b *Bot) buildMsg() string {
	msgs := []string{messageHeader}
	for _, keyId := range b.VeIDApiKey {
		ss := strings.Split(keyId, "@@")
		// api request
		resp, err := resty.New().R().
			SetResult(&VPSInfo{}).
			ForceContentType("application/json").
			SetQueryParam("veid", ss[0]).
			SetQueryParam("api_key", ss[1]).
			Get(apiAddress)

		if err != nil {
			logrus.Errorf("[%v] request failed: %v", ss[0], err)
			continue
		}

		rs := resp.Result().(*VPSInfo)
		msgs = append(msgs, fmt.Sprintf(messageTpl,
			fmt.Sprintf("`%s`", strings.Replace(strings.Split(rs.NodeDatacenter, ",")[0], ":", " -", -1)),
			fmt.Sprintf("`%sG / %sG`", fmt.Sprintf("%0.2f", float64(rs.VeUsedDiskSpaceB)/1024/1024/1024), rs.VeDiskQuotaGb),
			fmt.Sprintf("`%sG / %dG`", fmt.Sprintf("%0.2f", float64(rs.DataCounter)/1024/1024/1024), rs.PlanMonthlyData/1024/1024/1024),
			fmt.Sprintf("`%s`", time.Unix(rs.DataNextReset, 0).Format("2006-01-02 15:04:05")),
		))
	}

	return strings.Join(msgs, "\n")
}

func (b *Bot) Send() {
	msg := b.buildMsg()
	for _, id := range b.ChartIds {
		_, err := b.bot.Send(telebot.ChatID(id), msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		if err != nil {
			logrus.Errorf("failed to send notifaction message: %v", err)
		}
	}
}

func (b *Bot) Init() {
	tb, err := telebot.NewBot(telebot.Settings{
		Token:  b.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		logrus.Fatalf("failed to create telegram bot client: %v", err)
	}

	b.bot = tb

	b.bot.Handle("/query", func(ctx telebot.Context) error {
		_, err := b.bot.Send(ctx.Sender(), b.buildMsg(), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		return err
	})
}

func (b *Bot) Stop() {
	b.bot.Stop()
}