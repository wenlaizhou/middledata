package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wenlaizhou/middleware"
	_ "net/http/pprof"
	"strings"
	"time"
)

func init() {

}

const DSN = "%s:%s@tcp(%s)/%s"

var Config middleware.Config

var logger = middleware.GetLogger("middledata")

func Start(configPath string) {

	Config = middleware.LoadConfig(configPath)

	if Config == nil {
		logger.ErrorF("配置文件错误")
		return
	}

	logger.Info(middleware.ConfPrint(Config))

	dsn := fmt.Sprintf(DSN,
		middleware.ConfUnsafe(Config, "username"),
		middleware.ConfUnsafe(Config, "password"),
		middleware.ConfUnsafe(Config, "host"),
		middleware.ConfUnsafe(Config, "database"),
	)
	DB, err := sql.Open(middleware.ConfUnsafe(Config, "driver"), dsn)
	if err != nil {
		logger.Error(fmt.Sprintf("Open mysql failed,err:%v", err))
		return
	}
	DB.SetConnMaxLifetime(100 * time.Second) // 最大连接周期，超过时间的连接就close
	DB.SetMaxOpenConns(10)                   // 设置最大连接数
	DB.SetMaxIdleConns(2)                    // 设置闲置连接数

	SqlConf := map[string]interface{}{}

	err = json.Unmarshal([]byte(middleware.ReadString(middleware.ConfUnsafe(Config, "sql"))), &SqlConf)
	if err != nil {
		logger.Error(fmt.Sprintf("sql conf error : %v;", err))
		return
	}

	middleware.RegisterHandler("metrics", func(context middleware.Context) {
		var res []string
		for k, v := range SqlConf {
			data := v.(map[string]interface{})
			rows, err := DB.Query(fmt.Sprintf("%v", data["sql"]))
			if err != nil {
				logger.ErrorF("%v", err)
			}
			value := ""
			if rows.Next() {
				var x []uint8
				err = rows.Scan(&x)
				value = fmt.Sprintf("%s", x)
			}
			metricsData := middleware.MetricsData{
				Key:   k,
				Value: value,
				Tags:  map[string]string{},
			}
			for k, v := range data {
				if k == "sql" {
					continue
				}
				metricsData.Tags[k] = fmt.Sprintf("%v", v)
			}
			res = append(res, middleware.FormatMetricsData(metricsData))
		}

		context.OK("text/plain", []byte(strings.Join(res, "\n")))
	})

	middleware.StartServer("0.0.0.0", middleware.ConfIntUnsafe(Config, "port"))
}

func main() {

	Start("app.properties")

}
