package forest

import "github.com/go-xorm/xorm"

type MysqlDB struct {
    DbUrl  string
    engine *xorm.Engine
}

func (m *MysqlDB) Insert(snapshot *JobExecuteSnapshot) error {
    _, err := m.engine.Insert(snapshot)
    if err != nil {
        return err
    }
    return nil
}

func (m *MysqlDB) Update(snapshot *JobExecuteSnapshot) error {
    _, _ = m.engine.Where("id=?", snapshot.Id).Cols("status", "finish_time", "times", "result").Update(snapshot)
    return nil
}

func (m *MysqlDB) Get(snapshot *JobExecuteSnapshot) (exist bool, err error) {
    exist, err = m.engine.Where("id=?", snapshot.Id).Get(snapshot)
    if err != nil {
        return
    }
    return exist, err
}

func (m *MysqlDB) List(query *QueryExecuteSnapshotParam) (snapshots []*JobExecuteSnapshot, count int64, totalPage int64, err error) {
    var (
        where      *xorm.Session
        queryWhere *xorm.Session
    )
    snapshots = make([]*JobExecuteSnapshot, 0)
    where = m.engine.Where("1=1")
    queryWhere = m.engine.Where("1=1")
    if query.Id != "" {
        where.And("id=?", query.Id)
        queryWhere.And("id=?", query.Id)
    }
    if query.Group != "" {
        where.And("`group`=?", query.Group)
        queryWhere.And("`group`=?", query.Group)
    }

    if query.Ip != "" {

        where.And("ip=?", query.Ip)
        queryWhere.And("ip=?", query.Ip)
    }
    if query.Name != "" {
        where.And("name=?", query.Name)
        queryWhere.And("name=?", query.Name)
    }
    if query.Status != 0 {
        where.And("`status`=?", query.Status)
        queryWhere.And("`status`=?", query.Status)
    }
    if count, err = where.Count(&JobExecuteSnapshot{}); err != nil {
        return
    }

    if count > 0 {
        err = queryWhere.Desc("create_time").Limit(query.PageSize, (query.PageNo-1)*query.PageSize).Find(&snapshots)
        if err != nil {
            return
        }

        if count%int64(query.PageSize) == 0 {
            totalPage = count / int64(query.PageSize)
        } else {
            totalPage = count/int64(query.PageSize) + 1
        }

    }
    return
}

func (m *MysqlDB) Init() error {
    engine, err := xorm.NewEngine("mysql", m.DbUrl)
    if err != nil {
        return err
    }

    m.engine = engine
    return err
}
