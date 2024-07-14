package ioc

import (
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/user/internal/repository/dao"
	promsdk "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"time"
)

// InitDB 初始化数据库连接
func InitDB(l accesslog.Logger) *gorm.DB {
	// username:password@protocol(address)/dbname
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var config Config
	err := viper.UnmarshalKey("db", &config)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(config.DSN), &gorm.Config{
		// 配置logger
		Logger: glogger.New(gormWriterFunc(l.Debug), glogger.Config{
			// 慢查询阈值，只有执行时间超过这个阈值，才会使用
			// 50ms， 100ms
			// SQL 查询必然要求命中索引，最好就是走一次磁盘 IO
			// 一次磁盘 IO 是不到 10ms
			SlowThreshold:             time.Millisecond * 20,
			IgnoreRecordNotFoundError: true,
			// 参数查询
			ParameterizedQueries: true,
			LogLevel:             glogger.Info,
		}),
	})
	if err != nil {
		// panic 相当于goroutine 结束
		panic(err)
	}

	// 统计性能开销
	err = db.Use(prometheus.New(prometheus.Config{
		DBName: "webook",
		// 每15秒采集
		RefreshInterval: 60,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	// 监控查询的执行时间
	pcb := newCallbacks()
	// 注册插件
	db.Use(pcb)

	db.Use(tracing.NewPlugin(tracing.WithDBName("webook"),
		tracing.WithQueryFormatter(func(query string) string {
			l.Debug("", accesslog.String("query", query))
			return query
		}),
		// 不记录查询参数
		tracing.WithoutQueryVariables(),
		// 不记录metrics
		tracing.WithoutMetrics(),
	))

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	// 生成数据表结构
	return db
}

// 使用适配器实现  gorm的Writer 接口
type gormWriterFunc func(msg string, args ...accesslog.Field)

func (g gormWriterFunc) Printf(msg string, args ...interface{}) {
	// 调用本身的方法
	g(msg, accesslog.Field{Key: "arges", Value: args})
}

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func (c *Callbacks) Name() string {
	return "prometheus-query"
}

func (c *Callbacks) Initialize(db *gorm.DB) error {
	c.registerAll(db)
	return nil
}

func newCallbacks() *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		Namespace: "qinye_yiyi",
		Subsystem: "demo_user",
		Name:      "gorm_query_time",
		Help:      "统计 GORM 执行时间",
		ConstLabels: map[string]string{
			"db": "demo",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})
	cb := &Callbacks{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return cb
}

func (c *Callbacks) registerAll(db *gorm.DB) {
	db.Callback().Create().Before("*").
		Register("promethues_create_before", c.before())
	db.Callback().Create().After("*").
		Register("promethues_create_after", c.after("create"))

	db.Callback().Update().Before("*").
		Register("promethues_update_before", c.before())
	db.Callback().Update().After("*").
		Register("promethues_update_after", c.after("update"))

	db.Callback().Delete().Before("*").
		Register("promethues_delete_before", c.before())
	db.Callback().Delete().After("*").
		Register("promethues_delete_after", c.after("delete"))

	db.Callback().Raw().Before("*").
		Register("promethues_raw_before", c.before())
	db.Callback().Raw().After("*").
		Register("promethues_raw_after", c.after("raw"))

	db.Callback().Row().Before("*").
		Register("promethues_row_before", c.before())
	db.Callback().Row().After("*").
		Register("promethues_row_after", c.after("row"))
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		// 类型断言
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		// 上报prometheus
		c.vector.WithLabelValues(typ, table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}
