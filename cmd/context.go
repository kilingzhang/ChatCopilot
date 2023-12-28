package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lw396/WeComCopilot/pkg/db"
	"github.com/lw396/WeComCopilot/pkg/log"
	"github.com/lw396/WeComCopilot/pkg/redis"
	"github.com/lw396/WeComCopilot/pkg/snowflake"
	"github.com/lw396/WeComCopilot/pkg/valuer"
	"github.com/lw396/WeComCopilot/pkg/wechat"

	"github.com/silenceper/wechat/v2/cache"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
)

type Context struct {
	*cli.Context
	cfg         *ini.File
	appName     string
	environment string
	podId       uint
}

func (c *Context) Section(name string) *ini.Section {
	return c.cfg.Section(name)
}

func buildContext(c *cli.Context, appName string) (*Context, error) {
	environment := getEnv()
	name := strings.ToLower(appName)
	configDir := c.String("config-dir")

	logger := log.NewConsoleLogger("LOAD")
	logger.Infof("当前环境: %s", environment)
	logger.Infof("当前应用: %s", name)
	logger.Infof("配置目录: %s", configDir)
	logger.Info("开始加载配置...")

	fileNames := []string{
		"app.cfg",
		fmt.Sprintf("app.%s.cfg", environment),
		fmt.Sprintf("%s.cfg", name),
		fmt.Sprintf("%s.%s.cfg", name, environment),
	}

	var sources []interface{}
	for _, fileName := range fileNames {
		logger.Infof("加载配置文件: %s", fileName)
		sources = append(sources, filepath.Join(configDir, fileName))
	}

	opt := ini.LoadOptions{
		Loose:                   true,
		SkipUnrecognizableLines: true,
	}

	cfg := ini.Empty(opt)
	if len(sources) > 0 {
		var err error
		cfg, err = ini.LoadSources(opt, sources[0], sources[1:]...)
		if err != nil {
			return nil, err
		}
	}

	return &Context{
		Context:     c,
		cfg:         cfg,
		appName:     name,
		environment: environment,
		podId:       c.Uint("pod-id"),
	}, nil
}

func getEnv() string {
	environment := strings.ToLower(os.Getenv("METAPLASIA_ENV"))

	if environment == "" {
		environment = "develop"
	}
	return environment
}

func (c *Context) IsDebug() bool {
	return c.environment == "develop"
}

func (c *Context) buildLogger(scope string) log.Logger {
	if c.IsDebug() {
		return log.NewConsoleLogger(scope)
	}

	return log.NewLogger(log.Config{
		App:    c.appName,
		Scope:  scope,
		LogDir: c.String("log-dir"),
	})
}

func (c *Context) buildDB() (*gorm.DB, error) {
	host := valuer.Value("127.0.0.1").Try(
		os.Getenv("MYSQL_HOST"),
		c.Section("mysql").Key("host").String(),
	).String()
	port := valuer.Value(3306).Try(
		os.Getenv("MYSQL_PORT"),
		c.Section("mysql").Key("port").MustInt(),
	).Int()
	name := valuer.Value("github.com/lw396/WeComCopilot").Try(
		os.Getenv("MYSQL_DB"),
		c.Section("mysql").Key("db").String(),
	).String()
	user := valuer.Value("root").Try(
		os.Getenv("MYSQL_USER"),
		c.Section("mysql").Key("user").String(),
	).String()
	password := valuer.Value("secret").Try(
		os.Getenv("MYSQL_PASSWORD"),
		c.Section("mysql").Key("password").String(),
	).String()
	timezone := valuer.Value("UTC").Try(
		os.Getenv("MYSQL_TIMEZONE"),
		c.Section("mysql").Key("timezone").String(),
	).String()

	loc := url.QueryEscape(timezone)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=%s",
		user,
		password,
		host,
		port,
		name,
		loc,
	)

	idGen, err := snowflake.NewWithConfig(snowflake.Config{
		StartTime:    1648684800000,
		WorkerIDBits: 5,
		SequenceBits: 12,
		WorkerID:     int(c.podId),
	})
	if err != nil {
		return nil, err
	}

	return db.New(
		db.WithDSN(dsn),
		db.WithIDGenerator(idGen),
		db.WithLogger(c.buildLogger("DB")),
	)
}

func (c *Context) buildRedis() (redis.RedisClient, error) {
	host := valuer.Value("127.0.0.1").Try(
		os.Getenv("REDIS_HOST"),
		c.Section("redis").Key("host").String(),
	).String()
	port := valuer.Value(6379).Try(
		os.Getenv("REDIS_PORT"),
		c.Section("redis").Key("port").MustInt(),
	).Int()
	password := valuer.Value("secret").Try(
		os.Getenv("REDIS_AUTH"),
		c.Section("redis").Key("auth").String(),
	).String()
	db := valuer.Value(0).Try(
		os.Getenv("REDIS_DB"),
		c.Section("redis").Key("db").MustInt(),
	).Int()

	return redis.NewClient(
		redis.WithAddress(host, port),
		redis.WithAuth("", password),
		redis.WithDB(db),
	)
}

func (c *Context) buildWechat() (wechat.WechatClient, error) {
	host := valuer.Value("127.0.0.1").Try(
		os.Getenv("REDIS_HOST"),
		c.Section("redis").Key("host").String(),
	).String()
	password := valuer.Value("secret").Try(
		os.Getenv("REDIS_AUTH"),
		c.Section("redis").Key("auth").String(),
	).String()
	db := valuer.Value(0).Try(
		os.Getenv("REDIS_DB"),
		c.Section("redis").Key("db").MustInt(),
	).Int()

	appId := valuer.Value("").Try(
		os.Getenv("WECHAT_APP_ID"),
		c.Section("wechat").Key("app-id").String(),
	).String()
	appSecret := valuer.Value("").Try(
		os.Getenv("WECHAT_APP_SECRET"),
		c.Section("wechat").Key("app-secret").String(),
	).String()
	token := valuer.Value("").Try(
		os.Getenv("WECHAT_TOKEN"),
		c.Section("wechat").Key("token").String(),
	).String()
	encodingAesKey := valuer.Value("").Try(
		os.Getenv("WECHAT_ENCODING_AES_KEY"),
		c.Section("wechat").Key("encoding-aes-key").String(),
	).String()

	redis := cache.RedisOpts{
		Host:        host,
		Password:    password,
		Database:    db,
		MaxIdle:     100,
		MaxActive:   10,
		IdleTimeout: 20,
	}
	return wechat.NewOfficialAccount(offConfig.Config{
		AppID:          appId,
		AppSecret:      appSecret,
		Token:          token,
		EncodingAESKey: encodingAesKey,
		Cache:          cache.NewRedis(context.Background(), &redis),
	})
}
