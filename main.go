package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var commit string
var crontab string
var bot Bot

var rootCmd = &cobra.Command{
	Use: "bandwagonmon",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Infof("搬瓦工机器人启动, Git Commit: %s", commit)
		logrus.Info("初始化 Telegram 机器人...")
		bot.Init()

		logrus.Info("创建定时推送任务...")
		c := cron.New()
		_, err := c.AddFunc(crontab, bot.Send)
		if err != nil {
			logrus.Fatalf("failed to create crontab instence: %v", err)
		}
		c.Start()

		logrus.Info("搬瓦工机器人启动成功...")

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		<-ctx.Done()

		c.Stop()
		logrus.Info("关闭定时任务...")

		bot.Stop()
		logrus.Info("关闭 Telegram 机器人...")

		logrus.Info("搬瓦工机器人已停止!")
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
	rootCmd.PersistentFlags().StringVar(&bot.Token, "telegram-bot-token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram Bot Token")
	rootCmd.PersistentFlags().Int64SliceVar(&bot.ChartIds, "telegram-chart-id", int64Slice("TELEGRAM_CHART_ID"), "Telegram Chat IDs")
	rootCmd.PersistentFlags().StringSliceVar(&bot.VeIDApiKey, "vps-veid-apikey", stringSlice("VPS_VEID_APIKEY"), "BandwagonHost VPS VeID And API Key(format: veid@@apikey)")
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
