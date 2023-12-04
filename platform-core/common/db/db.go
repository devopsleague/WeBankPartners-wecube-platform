package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/WeBankPartners/wecube-platform/platform-core/common/log"
	"github.com/WeBankPartners/wecube-platform/platform-core/models"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"xorm.io/core"
	"xorm.io/xorm"
	xorm_log "xorm.io/xorm/log"
)

const DBTransactionId = "transactionId"

var (
	MysqlEngine *xorm.Engine
)

func InitDatabase() error {
	connStr := fmt.Sprintf("%s:%s@%s(%s)/%s?collation=utf8mb4_unicode_ci&allowNativePasswords=true",
		models.Config.Database.User, models.Config.Database.Password, "tcp", fmt.Sprintf("%s:%s", models.Config.Database.Server, models.Config.Database.Port), models.Config.Database.DataBase)
	engine, err := xorm.NewEngine("mysql", connStr)
	if err != nil {
		log.Logger.Error("Init database connect fail", log.Error(err))
		return err
	}
	engine.SetMaxIdleConns(models.Config.Database.MaxIdle)
	engine.SetMaxOpenConns(models.Config.Database.MaxOpen)
	engine.SetConnMaxLifetime(time.Duration(models.Config.Database.Timeout) * time.Second)
	engine.SetTZDatabase(time.UTC)
	engine.SetTZLocation(time.UTC)
	if models.Config.Log.DbLogEnable {
		engine.SetLogger(&dbContextLogger{LogLevel: 1, ShowSql: true, Logger: log.DatabaseLogger})
	}
	// 使用驼峰式映射
	engine.SetMapper(core.SnakeMapper{})
	MysqlEngine = engine
	if err = CheckDbConnection(); err == nil {
		log.Logger.Info("Success init database connect !!")
	} else {
		log.Logger.Error("Init database fail", log.Error(err))
	}
	return err
}

type dbContextLogger struct {
	LogLevel xorm_log.LogLevel
	ShowSql  bool
	Logger   *zap.Logger
}

func (d *dbContextLogger) BeforeSQL(ctx xorm_log.LogContext) {
}

func (d *dbContextLogger) AfterSQL(ctx xorm_log.LogContext) {
	t := ctx.Ctx.Value(DBTransactionId)
	var transactionId string
	if tmpTransactionId, ok := t.(string); ok {
		transactionId = tmpTransactionId
	}
	var costMs float64 = 0
	costTime := fmt.Sprintf("%s", ctx.ExecuteTime)
	if strings.Contains(costTime, "µs") {
		costMs, _ = strconv.ParseFloat(strings.ReplaceAll(costTime, "µs", ""), 64)
		costMs = costMs / 1000
	} else if strings.Contains(costTime, "ms") {
		costMs, _ = strconv.ParseFloat(costTime[:len(costTime)-2], 64)
	} else if strings.Contains(costTime, "s") && !strings.Contains(costTime, "m") {
		costMs, _ = strconv.ParseFloat(costTime[:len(costTime)-1], 64)
		costMs = costMs * 1000
	} else {
		costTime = costTime[:len(costTime)-1]
		mIndex := strings.Index(costTime, "m")
		minTime, _ := strconv.ParseFloat(costTime[:mIndex], 64)
		secTime, _ := strconv.ParseFloat(costTime[mIndex+1:], 64)
		costMs = (minTime*60 + secTime) * 1000
	}
	d.Logger.Info("["+transactionId+"]", log.String("sql", ctx.SQL), log.String("param", fmt.Sprintf("%v", ctx.Args)), log.Float64("cost_ms", costMs))
}

func (d *dbContextLogger) Debugf(format string, v ...interface{}) {
	d.Logger.Debug(fmt.Sprintf(format, v...))
}

func (d *dbContextLogger) Errorf(format string, v ...interface{}) {
	d.Logger.Debug(fmt.Sprintf(format, v...))
}

func (d *dbContextLogger) Infof(format string, v ...interface{}) {
	d.Logger.Debug(fmt.Sprintf(format, v...))
}

func (d *dbContextLogger) Warnf(format string, v ...interface{}) {
	d.Logger.Debug(fmt.Sprintf(format, v...))
}

func (d *dbContextLogger) Level() xorm_log.LogLevel {
	return d.LogLevel
}

func (d *dbContextLogger) SetLevel(l xorm_log.LogLevel) {
	d.LogLevel = l
}

func (d *dbContextLogger) ShowSQL(b ...bool) {
	d.ShowSql = b[0]
}

func (d *dbContextLogger) IsShowSQL() bool {
	return d.ShowSql
}

type ExecAction struct {
	Sql            string
	Param          []interface{}
	CheckAffectRow bool
}

func Transaction(actions []*ExecAction, ctx context.Context) error {
	if len(actions) == 0 {
		log.Logger.Warn("Transaction is empty,nothing to do")
		return fmt.Errorf("SQL exec transaction is empty,nothing to do,please check server log ")
	}
	for i, action := range actions {
		if action == nil {
			return fmt.Errorf("SQL exec transaction index%d is nill error,please check server log", i)
		}
	}
	session := MysqlEngine.NewSession().Context(ctx)
	err := session.Begin()
	for _, action := range actions {
		params := make([]interface{}, 0)
		params = append(params, action.Sql)
		for _, v := range action.Param {
			params = append(params, v)
		}

		execResult, execErr := session.Exec(params...)
		if execErr == nil && action.CheckAffectRow {
			if rowAffectNum, _ := execResult.RowsAffected(); rowAffectNum == 0 {
				execErr = fmt.Errorf("row affect 0 with exec sql:%v ", params)
			}
		}
		if execErr != nil {
			err = execErr
			session.Rollback()
			break
		}
	}
	if err == nil {
		err = session.Commit()
	}
	session.Close()
	return err
}

func CreateListParams(inputList []string, prefix string) (filterSql string, filterParam []interface{}) {
	if len(inputList) > 0 {
		var specList []string
		for _, v := range inputList {
			specList = append(specList, "?")
			filterParam = append(filterParam, prefix+v)
		}
		filterSql = strings.Join(specList, ",")
	}
	return
}

func CombineDBSql(input ...interface{}) string {
	var buf strings.Builder
	fmt.Fprint(&buf, input...)
	return buf.String()
}

func CheckDbConnection() (err error) {
	_, err = MysqlEngine.QueryString("select 1=1")
	return err
}

func NewDBCtx(transactionId string) context.Context {
	return context.WithValue(context.Background(), DBTransactionId, transactionId)
}