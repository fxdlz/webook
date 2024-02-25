package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func NewMongoDBArticleDAO(node *snowflake.Node, col *mongo.Collection, liveCol *mongo.Collection) ArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		col:     col,
		liveCol: liveCol,
	}
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Id = m.node.Generate().Int64()
	art.Ctime = now
	art.Utime = now
	_, err := m.col.InsertOne(ctx, art)
	return art.Id, err
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{"id", art.Id}, bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("ID 不对或者创作者不对")
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var err error
	id := art.Id
	if art.Id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	publishArt := PublishedArticle(art)
	publishArt.Utime = now
	filter := bson.D{bson.E{Key: "id", Value: publishArt.Id}, bson.E{Key: "author_id", Value: publishArt.AuthorId}}
	_, err = m.liveCol.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: art},
		bson.E{Key: "$setOnInsert", Value: bson.D{bson.E{Key: "ctime", Value: now}}}},
		options.Update().SetUpsert(true))
	return art.Id, err
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{Key: "id", Value: id}, bson.E{Key: "author_id", Value: uid}}
	res, err := m.col.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "status", Value: status}, bson.E{Key: "utime", Value: now}}}})
	if err != nil {
		return err
	}
	if res.UpsertedCount == 0 {
		return errors.New("ID不对或者创作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "status", Value: status}, bson.E{Key: "utime", Value: now}}}})
	return err
}
