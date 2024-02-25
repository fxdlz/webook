package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type MongoDBTestSuite struct {
	suite.Suite
	col *mongo.Collection
}

func (s *MongoDBTestSuite) SetupSuite() {
	t := s.T()
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
	s.col = col

	manyRes, err := col.InsertMany(ctx, []any{
		Article{
			Id:       123,
			AuthorId: 123,
		},
		Article{
			Id:       234,
			AuthorId: 234,
		},
	})

	assert.NoError(s.T(), err)
	s.T().Log("插入数量", len(manyRes.InsertedIDs))
}

func (s *MongoDBTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)

	_, err = s.col.Indexes().DropAll(ctx)
	assert.NoError(s.T(), err)
}

func (s *MongoDBTestSuite) TestOr() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	or := bson.A{bson.D{bson.E{Key: "id", Value: 123}}, bson.D{bson.E{Key: "id", Value: 234}}}
	res, err := s.col.Find(ctx, bson.D{bson.E{Key: "$or", Value: or}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestAnd() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	and := bson.A{bson.D{bson.E{Key: "id", Value: 123}}, bson.D{bson.E{Key: "author_id", Value: 123}}}
	res, err := s.col.Find(ctx, bson.D{bson.E{Key: "$and", Value: and}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestIn() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.D{bson.E{Key: "id", Value: bson.D{bson.E{Key: "$in", Value: []int{123, 234}}}}}
	proj := bson.M{"id": 1}
	res, err := s.col.Find(ctx, filter, options.Find().SetProjection(proj))
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := s.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_id"),
	})

	assert.NoError(s.T(), err)
	s.T().Log("创建索引", res)
}

func TestMongoDBQueries(t *testing.T) {
	suite.Run(t, &MongoDBTestSuite{})
}
