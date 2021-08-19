package forest

import (
    "context"
    "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
    mgoc "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
)

type MongoDB struct {
    Engine *mongo.Collection
}

func (m *MongoDB) Insert(snapshot *JobExecuteSnapshot) (err error) {
    coll := m.Engine
    _, err = coll.InsertOne(context.Background(), snapshot)
    if err != nil {
        return
    }
    return
}

func (m *MongoDB) Update(snapshot *JobExecuteSnapshot) (err error) {
    coll := m.Engine
    filter := bson.M{
        "id": snapshot.Id,
    }
    update := bson.M{
        "status":      snapshot.Status,
        "finish_time": snapshot.FinishTime,
        "times":       snapshot.Times,
        "result":      snapshot.Result,
    }
    _, err = coll.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return
    }
    return
}

func (m *MongoDB) Get(snapshot *JobExecuteSnapshot) (exist bool, err error) {
    coll := m.Engine
    filter := bson.M{
        "id": snapshot.Id,
    }
    singleResult := coll.FindOne(context.Background(), filter)
    if err = singleResult.Err(); err != nil {
        return false, nil
    }

    return true, nil
}

func (m *MongoDB) List(query *QueryExecuteSnapshotParam) (snapshots []*JobExecuteSnapshot, count int64, totalPage int64, err error) {
    ctx := context.Background()
    coll := m.Engine
    filter := bson.M{

    }
    if query.Id != "" {
        filter["id"] = query.Id
    }
    if query.Group != "" {
        filter["group"] = query.Group
    }

    if query.Ip != "" {
        filter["ip"] = query.Ip
    }
    if query.Name != "" {
        filter["name"] = query.Name
    }
    if query.Status != 0 {
        filter["status"] = query.Status
    }

    count, err = coll.CountDocuments(ctx, filter)
    if err != nil {
        return
    }

    l := int64(query.PageSize)
    s := int64(query.PageNo-1) * l

    cursor, err := coll.Find(ctx, filter,
        &options.FindOptions{
            Limit: &l,
            Skip:  &s,
            Sort:  bson.D{{"create_time", -1}},
        })
    if err != nil {
        return
    }

    defer cursor.Close(ctx)
    for cursor.Next(ctx) {
        result := &JobExecuteSnapshot{}
        err := cursor.Decode(result)
        if err != nil {
            return nil,0,0,err
        }
        snapshots = append(snapshots, result)
    }

    if count%int64(query.PageSize) == 0 {
        totalPage = count / int64(query.PageSize)
    } else {
        totalPage = count/int64(query.PageSize) + 1
    }
    return
}

func (m *MongoDB) Init() error {
    appConf := &config.AppConf{
        MongoDB: config.MongoDB{
            Name:            "default",
            Host:            "127.0.0.1:27017",
            User:            "",
            Pwd:             "",
            MaxPoolSize:     20,
            MaxConnIdleTime: 10,
        },
    }
    mgoc.InitDefalut(&appConf.MongoDB)

    cli, err := mgoc.DefaultClient()
    if err != nil {
        log.Println("mgoc.DefaultClient err: ", err)
        return err
    }

    m.Engine = cli.DB("forest").Collection("job_execute_snapshot")
    return nil
}
