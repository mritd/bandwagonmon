package main

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/telebot.v3"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var crontab string
var telegramBotToken string
var telegramChartIds []int64
var vpsVeIDApiKey []string

func send(bot *telebot.Bot) {
	msgs := []string{messageHeader}
	for _, keyId := range vpsVeIDApiKey {
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

	notiMsg := strings.Join(msgs, "\n")
	logrus.Infof("send notification:\n%s\n", notiMsg)
	for _, id := range telegramChartIds {
		_, err := bot.Send(telebot.ChatID(id), notiMsg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		if err != nil {
			logrus.Errorf("failed to send notifaction message: %v", err)
		}
	}
}

var rootCmd = &cobra.Command{
	Use: "bandwagonmon",
	Run: func(cmd *cobra.Command, args []string) {

		logrus.Info("init telegram bot client...")
		bot, err := telebot.NewBot(telebot.Settings{
			Token:  telegramBotToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		})
		if err != nil {
			logrus.Fatalf("failed to create telegram bot client: %v", err)
		}

		logrus.Info("create cron job...")
		c := cron.New()
		_, err = c.AddFunc(crontab, func() { send(bot) })
		if err != nil {
			logrus.Fatalf("failed to create crontab instence: %v", err)
		}
		c.Start()

		logrus.Info("listen http request...")
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`success`))
				return
			}
			send(bot)
		})

		srv := &http.Server{
			Addr:           "0.0.0.0:8080",
			Handler:        mux,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		go func() {
			_ = srv.ListenAndServe()
		}()

		logrus.Info("bandwagonmon started...")

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		<-ctx.Done()

		c.Stop()
		logrus.Info("cron job shutdown...")

		sctx, scancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer scancel()
		err = srv.Shutdown(sctx)
		if err != nil {
			logrus.Errorf("failed to shutdown http server: %v", err)
		} else {
			logrus.Info("http server shutdown...")
		}

		bot.Stop()
		logrus.Info("telegram bot shutdown...")

		logrus.Info("bandwagonmon stop!")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// CRON_TZ=Asia/Shanghai 30 12 * * *
	rootCmd.PersistentFlags().StringVar(&crontab, "crontab", os.Getenv("CRONTAB"), "Task Execution Crontab")
	rootCmd.PersistentFlags().StringVar(&telegramBotToken, "telegram-bot-token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram Bot Token")
	rootCmd.PersistentFlags().Int64SliceVar(&telegramChartIds, "telegram-chart-id", int64Slice("TELEGRAM_CHART_ID"), "Telegram Chat IDs")
	rootCmd.PersistentFlags().StringSliceVar(&vpsVeIDApiKey, "vps-veid-apikey", stringSlice("VPS_VEID_APIKEY"), "BandwagonHost VPS VeID And API Key(format: veid@@apikey)")
}

func int64Slice(key string) []int64 {
	var res []int64
	if os.Getenv(key) != "" {
		for _, s := range strings.Split(os.Getenv(key), ",") {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				logrus.Errorf("failed to parse int64: %v", err)
				continue
			}
			res = append(res, i)
		}
	}

	return res
}

func stringSlice(key string) []string {
	var res []string
	if os.Getenv(key) != "" {
		for _, s := range strings.Split(os.Getenv(key), ",") {
			res = append(res, s)
		}
	}

	return res
}
