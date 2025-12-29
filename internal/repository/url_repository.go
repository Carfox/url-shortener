package repository

import (
	"context"
	"fmt"
	"time"
	"url-shortener/internal/models"
	"url-shortener/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UrlRepository struct {
	Collection *mongo.Collection
}

func (r *UrlRepository) SaveUrl(url string) (models.UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		code := utils.CodeGenerator(6)
		filter := bson.M{"_id": code}
		err := r.Collection.FindOne(ctx, filter).Err()

		if err == nil {
			fmt.Println("Already exists a code saved", code)
			continue
		}

		if err != mongo.ErrNoDocuments {
			return models.UrlData{}, err
		}

		newURLRegister := models.UrlData{
			ID:        code,
			Url:       url,
			ShortCode: code,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, errInsert := r.Collection.InsertOne(ctx, newURLRegister)
		if errInsert != nil {
			return models.UrlData{}, errInsert
		}

		fmt.Println("URL Saved:", url, "->", code)
		return newURLRegister, nil
	}

}

func (r *UrlRepository) UpdateUrl(code string, newUrl string) (models.UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"_id": code}

	update := bson.M{
		"$set": bson.M{
			"url":        newUrl,
			"updated_at": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedDoc models.UrlData

	err := r.Collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDoc)

	return updatedDoc, err
}

func (r *UrlRepository) DeleteUrl(code string) {
	filter := bson.M{"_id": code}
	r.Collection.FindOneAndDelete(context.TODO(), filter)
}

func (r *UrlRepository) GetUrl(code string) (models.UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var doc models.UrlData

	filter := bson.M{"_id": code}

	err := r.Collection.FindOne(ctx, filter).Decode(&doc)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No documents found")
		} else {
			panic(err)
		}
	}

	return doc, err
}

func (r *UrlRepository) GetUrlAndIncrement(code string) (models.UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": code}
	update := bson.M{"$inc": bson.M{"access_count": 1}}

	var result models.UrlData
	err := r.Collection.FindOneAndUpdate(ctx, filter, update).Decode(&result)

	return result, err
}
