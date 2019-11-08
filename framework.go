package middledata

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wenlaizhou/middleware"
	"time"
)

func init() {

}

type Row map[string]interface{}

const DSN = "%s:%s@tcp(%s:%s)/%s"

var Config middleware.Config

var logger = middleware.GetLogger("middledata")

func Start(configPath string) {

	Config = middleware.LoadConfig(configPath)

	if Config == nil {
		logger.ErrorF("配置文件错误")
		return
	}

	dsn := fmt.Sprintf(DSN,
		middleware.ConfUnsafe(Config, "username"),
		middleware.ConfUnsafe(Config, "password"),
		middleware.ConfUnsafe(Config, "host"),
		middleware.ConfUnsafe(Config, "port"),
		middleware.ConfUnsafe(Config, "database"),
	)
	DB, err := sql.Open(middleware.ConfUnsafe(Config, "driver"), dsn)
	if err != nil {
		logger.Error(fmt.Sprintf("Open mysql failed,err:%v\n", err))
		return
	}
	DB.SetConnMaxLifetime(100 * time.Second) // 最大连接周期，超过时间的连接就close
	DB.SetMaxOpenConns(10)                   // 设置最大连接数
	DB.SetMaxIdleConns(2)                    // 设置闲置连接数

	middleware.RegisterHandler("metrics", func(context middleware.Context) {

	})

	middleware.StartServer("0.0.0.0", middleware.ConfIntUnsafe(Config, "port"))
}
