package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"note_service/internal/apperror"
	"note_service/internal/config"
	handlermodel "note_service/internal/handlers/notes"
	tagmodel "note_service/internal/handlers/tags"
	"note_service/pkg/logging"
)

type Storage struct {
	client         *mongo.Client
	noteCollection *mongo.Collection
	tagCollection  *mongo.Collection
}

type mongoNote struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Header       string             `bson:"header,omitempty"`
	Body         string             `bson:"body,omitempty"`
	ShortBody    string             `bson:"short_body,omitempty"`
	CreatedDate  int64              `bson:"created_date,omitempty"`
	CategoryUuid string             `bson:"category_uuid,omitempty"`
	Tags         []string           `bson:"tags,omitempty"`
}

type mongoTag struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name,omitempty"`
}

func NewStorage(cfg *config.Config) (storage *Storage, err error) {
	logger := logging.GetLogger().With(
		"layer", "storage",
		"db", "mongodb",
		"host", cfg.Mongo.Host,
		"port", cfg.Mongo.Port,
		"database", cfg.Mongo.Database,
		"collection", cfg.Mongo.Collection,
	)
	clientOpts := options.Client().
		ApplyURI(fmt.Sprintf("%s://%s:%s", cfg.Mongo.Scheme, cfg.Mongo.Host, cfg.Mongo.Port)).
		SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	if cfg.Mongo.Username != "" {
		clientOpts.SetAuth(options.Credential{
			AuthSource: cfg.Mongo.AuthSource,
			Username:   cfg.Mongo.Username,
			Password:   cfg.Mongo.Password,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client *mongo.Client
	client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.Error("failed to connect mongo", "error", err)
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	storage = &Storage{
		client:         client,
		noteCollection: client.Database(cfg.Mongo.Database).Collection(cfg.Mongo.Collection),
		tagCollection:  client.Database(cfg.Mongo.Database).Collection("tags"),
	}

	_, err = storage.tagCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		logger.Error("failed to create tag name index", "error", err)
		return nil, fmt.Errorf("create tag name index: %w", err)
	}

	if err = storage.Ping(); err != nil {
		logger.Error("mongo ping failed", "error", err)
		return nil, err
	}

	logger.Info("mongo connected")
	return storage, nil
}

func (s *Storage) Ping() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = s.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}

	return nil
}

func (s *Storage) Create(note handlermodel.Note) (noteUUID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.InsertOneResult
	result, err = s.noteCollection.InsertOne(ctx, note)
	if err != nil {
		return "", fmt.Errorf("insert note: %w", err)
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("inserted id is not object id")
	}

	noteUUID = objectID.Hex()
	return noteUUID, nil
}

func (s *Storage) FindOne(noteUUID string) (note handlermodel.Note, err error) {
	objectID, err := primitive.ObjectIDFromHex(noteUUID)
	if err != nil {
		return handlermodel.Note{}, apperror.BadRequestError("invalid note uuid")
	}

	filter := bson.M{"_id": objectID}
	findOptions := options.FindOne().SetProjection(bson.M{"short_body": 0})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := s.noteCollection.FindOne(ctx, filter, findOptions)
	if err = result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return handlermodel.Note{}, apperror.NotFoundError("note not found")
		}
		return handlermodel.Note{}, fmt.Errorf("find note: %w", err)
	}

	var rawNote mongoNote
	if err = result.Decode(&rawNote); err != nil {
		return handlermodel.Note{}, fmt.Errorf("decode note: %w", err)
	}

	note = handlermodel.Note{
		Uuid:         noteUUID,
		Header:       rawNote.Header,
		Body:         rawNote.Body,
		CreatedDate:  rawNote.CreatedDate,
		CategoryUuid: rawNote.CategoryUuid,
		Tags:         rawNote.Tags,
	}

	return note, nil
}

func (s *Storage) FindByCategoryUUID(categoryUUID string) (notes []handlermodel.Note, err error) {
	filter := bson.M{"category_uuid": categoryUUID}
	findOptions := options.Find().SetSort(bson.D{{Key: "created_date", Value: -1}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cursor *mongo.Cursor
	cursor, err = s.noteCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find notes by category: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var rawNote mongoNote
		if err = cursor.Decode(&rawNote); err != nil {
			return nil, fmt.Errorf("decode note item: %w", err)
		}

		notes = append(notes, handlermodel.Note{
			Uuid:         rawNote.ID.Hex(),
			Header:       rawNote.Header,
			Body:         rawNote.Body,
			ShortBody:    rawNote.ShortBody,
			CreatedDate:  rawNote.CreatedDate,
			CategoryUuid: rawNote.CategoryUuid,
			Tags:         rawNote.Tags,
		})
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor notes by category: %w", err)
	}

	if len(notes) == 0 {
		return nil, apperror.NotFoundError("notes not found")
	}

	return notes, nil
}

func (s *Storage) Update(noteUUID string, note handlermodel.Note, tagsUpdate bool) (err error) {
	objectID, err := primitive.ObjectIDFromHex(noteUUID)
	if err != nil {
		return apperror.BadRequestError("invalid note uuid")
	}

	var payloadBytes []byte
	payloadBytes, err = json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshal update dto: %w", err)
	}

	updateBody := bson.M{}
	if err = bson.UnmarshalExtJSON(payloadBytes, true, &updateBody); err != nil {
		return fmt.Errorf("unmarshal update dto to bson: %w", err)
	}

	delete(updateBody, "uuid")
	delete(updateBody, "created_date")

	if note.Body == "" {
		delete(updateBody, "body")
		delete(updateBody, "short_body")
	}
	if note.Header == "" {
		delete(updateBody, "header")
	}
	if note.CategoryUuid == "" {
		delete(updateBody, "category_uuid")
	}
	if !tagsUpdate {
		delete(updateBody, "tags")
	}

	update := bson.M{"$set": updateBody}
	if tagsUpdate {
		updateBody["tags"] = note.Tags
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.UpdateResult
	result, err = s.noteCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}

	if result.MatchedCount == 0 {
		return apperror.NotFoundError("note not found")
	}

	return nil
}

func (s *Storage) Delete(noteUUID string) (err error) {
	objectID, err := primitive.ObjectIDFromHex(noteUUID)
	if err != nil {
		return apperror.BadRequestError("invalid note uuid")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.DeleteResult
	result, err = s.noteCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	if result.DeletedCount == 0 {
		return apperror.NotFoundError("note not found")
	}

	return nil
}

func (s *Storage) CreateTag(tag tagmodel.Tag) (tagUUID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.InsertOneResult
	result, err = s.tagCollection.InsertOne(ctx, tag)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", apperror.BadRequestError("tag already exists")
		}
		return "", fmt.Errorf("insert tag: %w", err)
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("inserted tag id is not object id")
	}

	tagUUID = objectID.Hex()
	return tagUUID, nil
}

func (s *Storage) FindTags(tagUUIDs []string) (tags []tagmodel.Tag, err error) {
	filter := bson.M{}
	if len(tagUUIDs) > 0 {
		objectIDs := make([]primitive.ObjectID, 0, len(tagUUIDs))
		for _, tagUUID := range tagUUIDs {
			objectID, convErr := primitive.ObjectIDFromHex(tagUUID)
			if convErr != nil {
				return nil, apperror.BadRequestError("invalid tag uuid")
			}
			objectIDs = append(objectIDs, objectID)
		}
		filter["_id"] = bson.M{"$in": objectIDs}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cursor *mongo.Cursor
	cursor, err = s.tagCollection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("find tags: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var rawTag mongoTag
		if err = cursor.Decode(&rawTag); err != nil {
			return nil, fmt.Errorf("decode tag item: %w", err)
		}

		tags = append(tags, tagmodel.Tag{
			Uuid: rawTag.ID.Hex(),
			Name: rawTag.Name,
		})
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor tags: %w", err)
	}

	if len(tags) == 0 {
		return nil, apperror.NotFoundError("tags not found")
	}

	return tags, nil
}

func (s *Storage) CheckTagsExist(tagUUIDs []string) (err error) {
	if len(tagUUIDs) == 0 {
		return nil
	}

	objectIDs := make([]primitive.ObjectID, 0, len(tagUUIDs))
	uniqueIDs := make(map[string]struct{}, len(tagUUIDs))
	for _, tagUUID := range tagUUIDs {
		if _, ok := uniqueIDs[tagUUID]; ok {
			continue
		}
		uniqueIDs[tagUUID] = struct{}{}

		objectID, convErr := primitive.ObjectIDFromHex(tagUUID)
		if convErr != nil {
			return apperror.BadRequestError("invalid tag uuid")
		}
		objectIDs = append(objectIDs, objectID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int64
	count, err = s.tagCollection.CountDocuments(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		return fmt.Errorf("count tags: %w", err)
	}

	if count != int64(len(objectIDs)) {
		return apperror.NotFoundError("tags not found")
	}

	return nil
}

func (s *Storage) DeleteTag(tagUUID string) (err error) {
	objectID, err := primitive.ObjectIDFromHex(tagUUID)
	if err != nil {
		return apperror.BadRequestError("invalid tag uuid")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.DeleteResult
	result, err = s.tagCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	if result.DeletedCount == 0 {
		return apperror.NotFoundError("tag not found")
	}

	_, err = s.noteCollection.UpdateMany(
		ctx,
		bson.M{"tags": tagUUID},
		bson.M{"$pull": bson.M{"tags": tagUUID}},
	)
	if err != nil {
		return fmt.Errorf("pull deleted tag from notes: %w", err)
	}

	return nil
}
