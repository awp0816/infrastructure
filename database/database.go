package database

import (
	"database/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type Database struct {
	Driver       string `yaml:"driver"`
	Dsn          string `yaml:"dsn"`
	Debug        bool   `yaml:"debug"`
	MaxIdleConns int    `yaml:"idleConn"`
	MaxOpenConns int    `yaml:"openConn"`
	MaxLifeTime  int    `yaml:"lifeTime"`
	AutoMigrate  bool   `yaml:"auto_migrate"`
}

func SetupDatabase(l *zap.Logger, inParams Database) (dbConn *gorm.DB, err error) {
	switch inParams.Driver {
	case "mysql":
		if dbConn, err = gorm.Open(mysql.Open(inParams.Dsn), &gorm.Config{
			//迁移时是否禁用外键约束
			DisableForeignKeyConstraintWhenMigrating: true,
			//命名策略
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `user`
			},
			Logger: logger.Default.LogMode(logger.Warn),
		}); err != nil {
			l.Error("创建mysql连接失败", zap.Error(err))
			return
		}
	case "sqlite":
		if dbConn, err = gorm.Open(sqlite.Open(inParams.Dsn), &gorm.Config{
			//迁移时是否禁用外键约束
			DisableForeignKeyConstraintWhenMigrating: true,
			//命名策略
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `user`
			},
			Logger: logger.Default.LogMode(logger.Warn),
		}); err != nil {
			l.Error("创建sqlite连接失败", zap.Error(err))
			return
		}
		if err = dbConn.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
			return
		}
	default:
		err = errors.Errorf("unsupported database:%s", inParams.Driver)
		return
	}

	dbAction(dbConn, l)

	var sqldb *sql.DB
	if sqldb, err = dbConn.DB(); err != nil {
		l.Error("获取连接池DB失败", zap.Error(err))
		return
	}
	if err = sqldb.Ping(); err != nil {
		l.Error("PING连接池DB失败", zap.Error(err))
		return
	}
	//设置空闲连接池中连接的最大数量
	sqldb.SetMaxIdleConns(inParams.MaxIdleConns)
	//设置打开数据库连接的最大数量
	sqldb.SetMaxOpenConns(inParams.MaxOpenConns)
	//设置打开数据库的连接最大存活时间
	sqldb.SetConnMaxLifetime(time.Second * time.Duration(inParams.MaxLifeTime))

	if inParams.Debug {
		dbConn = dbConn.Debug()
	}
	return
}

func dbAction(db *gorm.DB, l *zap.Logger) {
	_ = db.Callback().Create().After("gorm:after_create").Register("log", afterLog(l))
	_ = db.Callback().Query().After("gorm:after_query").Register("log", afterLog(l))
	_ = db.Callback().Delete().After("gorm:after_delete").Register("log", afterLog(l))
	_ = db.Callback().Update().After("gorm:after_update").Register("log", afterLog(l))
	_ = db.Callback().Raw().After("gorm:raw").Register("log", afterLog(l))
}

func afterLog(l *zap.Logger) func(*gorm.DB) {
	return func(db *gorm.DB) {
		err := db.Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			sqlStr := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
			l.Error("Sql执行出错", zap.Error(err), zap.String("error_sql", sqlStr))
			return
		}
	}
}
