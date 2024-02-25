package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}
	ops := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(ctx, ops)
	assert.NoError(t, err)
	col := client.Database("webook").Collection("articles")
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 1,
	})
	assert.NoError(t, err)
	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("插入ID", oid)

	filter := bson.D{bson.E{Key: "id", Value: 1}}
	findRes := col.FindOne(ctx, filter)
	if findRes.Err() == mongo.ErrNoDocuments {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	filter = bson.D{bson.E{Key: "id", Value: 1}}
	sets := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "新的标题"}}}
	updateRes, err := col.UpdateOne(ctx, filter, sets)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateRes.ModifiedCount)

	updateRes, err = col.UpdateMany(ctx, filter, bson.D{bson.E{Key: "$set", Value: Article{
		Content: "新的内容",
	}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateRes.ModifiedCount)

	deleteRes, err := col.DeleteMany(ctx, filter)
	assert.NoError(t, err)
	t.Log("删除文本数量", deleteRes.DeletedCount)
}

type Article struct {
	Id      int64  `bson:"id,omitempty"`
	Title   string `bson:"title,omitempty"`
	Content string `bson:"content,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
