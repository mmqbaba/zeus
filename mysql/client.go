package mysqlclient

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	conf "gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	zeusctx "gitlab.dg.com/BackEnd/deliver/tif/zeus/context"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/errors"
	lock "gitlab.dg.com/BackEnd/deliver/tif/zeus/lock/redis"
	tracing "gitlab.dg.com/BackEnd/deliver/tif/zeus/trace"
)

const driverName = "mysql"

type DataSource struct {
	DataSourceName  string
	DriverName      string
	MaxIdleConns    int
	MaxOpenConns    int
	TraceOnlyLogErr bool
}

var ConnectionMap = make(map[string]*sql.DB)
var InstanceMap = make(map[string]DataSource)

type DBSession struct {
	Name            string
	Conn            *sql.DB
	TX              *sql.Tx
	isTranc         bool
	traceOnlyLogErr bool
}

func ReloadConfig(cfg map[string]conf.MysqlDB) error {
	var tmpInstanceMap = make(map[string]DataSource)
	for instance, sqlconf := range cfg {
		tmpInstanceMap[instance] = New(&sqlconf)
	}

	InstanceMap = tmpInstanceMap

	for DataSourceName, dbConn := range ConnectionMap {
		fmt.Printf("Closing connection: %v\n", DataSourceName)
		dbConn.Close()
		delete(ConnectionMap, DataSourceName)
	}

	return nil
}

func GetSession(ctx context.Context, instance string) (*DBSession, error) {
	logger := zeusctx.ExtractLogger(ctx)
	v, ok := InstanceMap[instance]
	if !ok {
		return nil, errors.ECodeMysqlErr.ParseErr("unknown instance: " + instance)
	}

	dataSourceName := v.DataSourceName

	dbConn := ConnectionMap[dataSourceName]
	if dbConn != nil {
		return &DBSession{Conn: dbConn}, nil
	}

	dbConn, err := sql.Open(v.DriverName, dataSourceName)
	if err != nil {
		errMsg := fmt.Sprintf("invalid session %s, error %s", instance, err)
		logger.Error(errMsg)
		return nil, errors.ECodeMysqlErr.ParseErr(errMsg)
	}

	dbConn.SetMaxIdleConns(v.MaxIdleConns)
	dbConn.SetMaxOpenConns(v.MaxOpenConns)
	ConnectionMap[dataSourceName] = dbConn
	return &DBSession{Name: instance, Conn: dbConn, traceOnlyLogErr: v.TraceOnlyLogErr}, nil
}

func ConnectionRelease() {
	for DataSourceName, dbConn := range ConnectionMap {
		fmt.Printf("Closing connection: %v\n", DataSourceName)
		dbConn.Close()
		delete(ConnectionMap, DataSourceName)
	}
}

func New(sqlconf *conf.MysqlDB) DataSource {
	ds := DataSource{
		DataSourceName:  sqlconf.DataSourceName,
		DriverName:      "mysql",
		MaxIdleConns:    sqlconf.MaxIdleConns,
		MaxOpenConns:    sqlconf.MaxOpenConns,
		TraceOnlyLogErr: sqlconf.TraceOnlyLogErr,
	}
	if ds.MaxIdleConns == 0 {
		ds.MaxIdleConns = 32
	}
	if ds.MaxOpenConns == 0 {
		ds.MaxOpenConns = 1024
	}
	return ds
}

// DoTransactWithLock DoTransactWithLock
// 这里使用了redis锁
func (session *DBSession) DoTransactWithLock(ctx context.Context, key string, txFunc func(*DBSession) error) (err error) {
	logger := zeusctx.ExtractLogger(ctx)
	options := &lock.Options{
		//The maximum duration to lock a key is 10s
		LockTimeout: 10 * time.Second,
		//The number of time the acquisition of a lock will be retried
		RetryCount: 10,
	}
	rdc, err := zeusctx.ExtractRedis(ctx)
	if err != nil {
		logger.Error(err)
		return
	}
	// Obtain a new lock with options settings
	// options is nil if you will obtain a new lock with default settings
	lock, err := lock.Obtain(ctx, rdc, key, options)
	if err != nil {
		return
	} else if lock == nil {
		return errors.ECodeLockNotObtained.ParseErr("")
	}
	// Don't forget to unlock in the end
	defer func() {
		if p := recover(); p != nil {
			lock.Unlock()
			panic(p) // re-throw panic after unlock
		} else {
			lock.Unlock()
		}
	}()
	return session.DoTransact(ctx, txFunc)
}

func (session *DBSession) DoTransact(ctx context.Context, txFunc func(*DBSession) error) (err error) {
	logger := zeusctx.ExtractLogger(ctx)
	tx, err := session.Conn.Begin()
	logger.Infof("transaction begin: %+v \n", tx)
	if err != nil {
		return errors.ECodeMysqlErr.ParseErr("begin transaction error")
	}
	defer func() {
		session.isTranc = false
		if p := recover(); p != nil {
			tx.Rollback()
			logger.Infof("transaction rollback complete :%+v \n", tx)
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
			logger.Infof("transaction rollback complete:%+v \n", tx)
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			logger.Infof("transaction commit:%+v \n", tx)
		}
	}()
	session.TX = tx
	session.isTranc = true
	err = txFunc(session)
	return
}

func (session *DBSession) query(ctx context.Context, sqlStmt string, args ...interface{}) ([]map[string]interface{}, error) {
	var (
		rows *sql.Rows
		err  error
	)
	logger := zeusctx.ExtractLogger(ctx)
	if session.isTranc {
		rows, err = session.TX.Query(sqlStmt, args...)
	} else {
		rows, err = session.Conn.Query(sqlStmt, args...)
	}
	if err != nil {
		errMsg := fmt.Sprintf("invalid statement, session %s,  sql %s,error %s", session.Name, sqlStmt, err)
		logger.Error(errMsg)
		return nil, errors.ECodeMysqlErr.ParseErr(errMsg)
	}

	columns, err := rows.Columns()
	if err != nil {
		errMsg := fmt.Sprintf("invalid statement, session %s,  sql %s,error %s", session.Name, sqlStmt, err)
		logger.Error(errMsg)
		return nil, errors.ECodeMysqlErr.ParseErr(errMsg)
	}
	result := make([]map[string]interface{}, 0)
	values := make([]sql.NullString, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			errMsg := fmt.Sprintf("invalid statement, session %s,  sql %s,error %s", session.Name, sqlStmt, err)
			logger.Error(errMsg)
			return nil, errors.ECodeMysqlErr.ParseErr(errMsg)
		}

		data := make(map[string]interface{})
		for i, v := range values {
			if v.Valid {
				data[columns[i]] = v.String
			} else {
				data[columns[i]] = nil
			}
		}

		result = append(result, data)
	}
	return result, nil
}

func (session *DBSession) exeSql(ctx context.Context, sqlStmt string, args ...interface{}) (affected int64, rowid int64, err error) {
	logger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := "mysql.ExecSql"
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	defer func() {
		if session.traceOnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()
	ext.SpanKindConsumer.Set(span)
	span.SetTag("sqlstmt", sqlStmt)
	var res sql.Result
	if session.isTranc {
		res, err = session.TX.Exec(sqlStmt, args...)
	} else {
		res, err = session.Conn.Exec(sqlStmt, args...)
	}
	if err != nil {
		errMsg := fmt.Sprintf("invalid statement, session %s,  sql %s,error %s", session.Name, sqlStmt, err)
		logger.Error(errMsg)
		span.SetTag("result.error", errors.ECodeMysqlErr.ParseErr(errMsg))
		return -1, -1, errors.ECodeMysqlErr.ParseErr(errMsg)
	}

	affected, err = res.RowsAffected()
	if err != nil {
		errMsg := fmt.Sprintf("invalid statement, session %s,  sql %s,error %s", session.Name, sqlStmt, err)
		logger.Error(errMsg)
		span.SetTag("result.error", errors.ECodeMysqlErr.ParseErr(errMsg))
		return -1, -1, errors.ECodeMysqlErr.ParseErr(errMsg)
	}
	span.SetTag("result.affected", affected)
	rowid, err = res.LastInsertId()
	if err != nil {
		errMsg := fmt.Sprintf("invalid statement, session %s,  sql %s,error %s", session.Name, sqlStmt, err)
		logger.Error(errMsg)
		span.SetTag("result.error", errors.ECodeMysqlErr.ParseErr(errMsg))
		return -1, -1, errors.ECodeMysqlErr.ParseErr(errMsg)
	}
	span.SetTag("result.lastInsertId", rowid)

	return affected, rowid, nil
}

func (session *DBSession) ExecSql(ctx context.Context, sqlStmt string, args ...interface{}) (affected int64,
	err error) {
	affected, _, err = session.exeSql(ctx, sqlStmt, args...)
	if err != nil {
		return affected, err
	}
	return affected, nil
}

func (session *DBSession) ExecSqlWithIncrementIdReturn(ctx context.Context, sqlStmt string, args ...interface{}) (affected int64, rowid int64,
	err error) {
	return session.exeSql(ctx, sqlStmt, args...)
}

func (session *DBSession) Count(ctx context.Context, sqlStmt string, args ...interface{}) (count int,
	err error) {
	logger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := "mysql.Count"
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	defer func() {
		if session.traceOnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()
	ext.SpanKindConsumer.Set(span)
	span.SetTag("sqlstmt", sqlStmt)
	result, err := session.query(ctx, sqlStmt, args...)
	if err != nil {
		logger.Error(err)
		span.SetTag("result.error", err)
		return
	}
	if len(result) != 1 {
		errMsg := fmt.Sprintf("invalid statement, result len %d, session %s,  sql %s", len(result), session.Name, sqlStmt)
		logger.Error(errMsg)
		span.SetTag("result.error", errors.ECodeMysqlErr.ParseErr(errMsg))
		err = errors.ECodeMysqlErr.ParseErr(errMsg)
		return
	}
	r := result[0]
	if len(r) != 1 {
		errMsg := fmt.Sprintf("invalid statement, result len %d, session %s,  sql %s", len(r), session.Name, sqlStmt)
		logger.Error(errMsg)
		span.SetTag("result.error", errors.ECodeMysqlErr.ParseErr(errMsg))
		err = errors.ECodeMysqlErr.ParseErr(errMsg)
		return
	}

	var reason string
	for _, val := range r {
		if reflect.TypeOf(val).Kind() == reflect.String {
			var i int
			i, err := strconv.Atoi(val.(string))
			if err == nil {
				span.SetTag("result.affected", val.(string))
				return i, nil
			} else {
				reason = fmt.Sprintf("not integer [%s]", val.(string))
			}
		} else {
			reason = fmt.Sprintf("not string type [%s]",
				reflect.TypeOf(val).Name())
		}
		break
	}
	errMsg := fmt.Sprintf("invalid statement, result len %d, session %s, sql %s,reason %s", len(r), session.Name, sqlStmt, reason)
	logger.Error(errMsg)
	span.SetTag("result.error", errors.ECodeMysqlErr.ParseErr(errMsg))
	err = errors.ECodeMysqlErr.ParseErr(errMsg)
	return
}

func (session *DBSession) QuerySql(ctx context.Context, sqlStmt string,
	args ...interface{}) (result []map[string]interface{}, err error) {
	logger := zeusctx.ExtractLogger(ctx)
	tracer := tracing.NewTracerWrap(opentracing.GlobalTracer())
	name := "mysql.QuerySql"
	ctx, span, _ := tracer.StartSpanFromContext(ctx, name)
	defer func() {
		if session.traceOnlyLogErr && err == nil {
			return
		}
		span.Finish()
	}()
	ext.SpanKindConsumer.Set(span)
	span.SetTag("sqlstmt", sqlStmt)
	result, err = session.query(ctx, sqlStmt, args...)
	if err != nil {
		logger.Error(err)
		span.SetTag("result.error", err)
		return nil, err
	}
	span.SetTag("result.affected", len(result))
	return result, nil
}
