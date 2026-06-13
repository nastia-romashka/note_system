package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	"note_service/internal/storage"
	"note_service/pkg/logging"
)

type Storage struct {
	client         *mongo.Client
	noteCollection *mongo.Collection
	tagCollection  *mongo.Collection
}

type mongoNote struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserUuid       string             `bson:"user_uuid,omitempty"`
	WorkspaceID    string             `bson:"workspace_id,omitempty"`
	AuthorUserUUID string             `bson:"author_user_uuid,omitempty"`
	Header         string             `bson:"header,omitempty"`
	Body           string             `bson:"body,omitempty"`
	ShortBody      string             `bson:"short_body,omitempty"`
	CreatedDate    int64              `bson:"created_date,omitempty"`
	UpdatedAt      int64              `bson:"updated_at,omitempty"`
	CategoryUuid   string             `bson:"category_uuid,omitempty"`
	Tags           []string           `bson:"tags,omitempty"`
	Event          *mongoNoteEvent    `bson:"event,omitempty"`
}

type mongoNoteEvent struct {
	Enabled bool  `bson:"enabled"`
	StartAt int64 `bson:"start_at,omitempty"`
	EndAt   int64 `bson:"end_at,omitempty"`
}

type mongoTag struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserUuid    string             `bson:"user_uuid,omitempty"`
	WorkspaceID string             `bson:"workspace_id,omitempty"`
	Name        string             `bson:"name,omitempty"`
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

	_, _ = storage.tagCollection.Indexes().DropOne(ctx, "name_1")

	_, err = storage.tagCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().
				SetName("workspace_id_1_name_1").
				SetUnique(true),
		},
	)
	if err != nil {
		logger.Error("failed to create workspace tag name index", "error", err)
		return nil, fmt.Errorf("create workspace tag name index: %w", err)
	}

	_, err = storage.noteCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "workspace_id", Value: 1},
				{Key: "category_uuid", Value: 1},
				{Key: "created_date", Value: -1},
			},
			Options: options.Index().SetName("workspace_id_1_category_uuid_1_created_date_-1"),
		},
	)
	if err != nil {
		logger.Error("failed to create workspace note category index", "error", err)
		return nil, fmt.Errorf("create workspace note category index: %w", err)
	}

	_, err = storage.noteCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "workspace_id", Value: 1},
				{Key: "event.enabled", Value: 1},
				{Key: "event.start_at", Value: 1},
			},
			Options: options.Index().SetName("workspace_id_1_event.enabled_1_event.start_at_1"),
		},
	)
	if err != nil {
		logger.Error("failed to create workspace note calendar index", "error", err)
		return nil, fmt.Errorf("create workspace note calendar index: %w", err)
	}

	_, err = storage.noteCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "workspace_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("workspace_id_1_updated_at_-1"),
		},
	)
	if err != nil {
		logger.Error("failed to create workspace note updated_at index", "error", err)
		return nil, fmt.Errorf("create workspace note updated_at index: %w", err)
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

func noteScopeFilter(scope storage.Scope) bson.M {
	return bson.M{"workspace_id": scope.WorkspaceID}
}

func noteAccessFilter(scope storage.Scope) bson.M {
	return noteScopeFilter(scope)
}

func tagScopeFilter(scope storage.Scope) bson.M {
	return bson.M{"workspace_id": scope.WorkspaceID}
}

func noteScopeWithTagFilter(scope storage.Scope, tagUUID string) bson.M {
	filter := noteScopeFilter(scope)
	filter["tags"] = tagUUID
	return filter
}

func (s *Storage) FindOne(noteUUID string, scope storage.Scope) (note handlermodel.Note, err error) {
	objectID, err := primitive.ObjectIDFromHex(noteUUID)
	if err != nil {
		return handlermodel.Note{}, apperror.BadRequestError("invalid note uuid")
	}

	filter := noteAccessFilter(scope)
	filter["_id"] = objectID
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
		Uuid:           noteUUID,
		UserUuid:       rawNote.UserUuid,
		WorkspaceID:    rawNote.WorkspaceID,
		AuthorUserUUID: rawNote.AuthorUserUUID,
		Header:         rawNote.Header,
		Body:           rawNote.Body,
		CreatedDate:    rawNote.CreatedDate,
		UpdatedAt:      rawNote.UpdatedAt,
		CategoryUuid:   rawNote.CategoryUuid,
		Tags:           rawNote.Tags,
		Event:          toHandlerEvent(rawNote.Event),
	}

	return note, nil
}

func (s *Storage) FindByCategoryUUID(categoryUUID string, scope storage.Scope) (notes []handlermodel.Note, err error) {
	notes = make([]handlermodel.Note, 0)

	filter := noteScopeFilter(scope)
	filter["category_uuid"] = categoryUUID
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
			Uuid:           rawNote.ID.Hex(),
			UserUuid:       rawNote.UserUuid,
			WorkspaceID:    rawNote.WorkspaceID,
			AuthorUserUUID: rawNote.AuthorUserUUID,
			Header:         rawNote.Header,
			Body:           rawNote.Body,
			ShortBody:      rawNote.ShortBody,
			CreatedDate:    rawNote.CreatedDate,
			UpdatedAt:      rawNote.UpdatedAt,
			CategoryUuid:   rawNote.CategoryUuid,
			Tags:           rawNote.Tags,
			Event:          toHandlerEvent(rawNote.Event),
		})
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor notes by category: %w", err)
	}

	return notes, nil
}

func (s *Storage) FindByEventRange(from, to int64, scope storage.Scope) (notes []handlermodel.Note, err error) {
	notes = make([]handlermodel.Note, 0)

	filter := noteScopeFilter(scope)
	filter["event.enabled"] = true
	filter["event.start_at"] = bson.M{"$gte": from, "$lte": to}
	findOptions := options.Find().SetSort(bson.D{{Key: "event.start_at", Value: 1}, {Key: "_id", Value: 1}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.noteCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find notes by event range: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var rawNote mongoNote
		if err = cursor.Decode(&rawNote); err != nil {
			return nil, fmt.Errorf("decode calendar note item: %w", err)
		}

		notes = append(notes, handlermodel.Note{
			Uuid:           rawNote.ID.Hex(),
			UserUuid:       rawNote.UserUuid,
			WorkspaceID:    rawNote.WorkspaceID,
			AuthorUserUUID: rawNote.AuthorUserUUID,
			Header:         rawNote.Header,
			Body:           rawNote.Body,
			ShortBody:      rawNote.ShortBody,
			CreatedDate:    rawNote.CreatedDate,
			UpdatedAt:      rawNote.UpdatedAt,
			CategoryUuid:   rawNote.CategoryUuid,
			Tags:           rawNote.Tags,
			Event:          toHandlerEvent(rawNote.Event),
		})
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor notes by event range: %w", err)
	}

	return notes, nil
}

func (s *Storage) CountStats(scope storage.Scope) (stats handlermodel.NoteStats, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats.NotesCount, err = s.noteCollection.CountDocuments(ctx, noteScopeFilter(scope))
	if err != nil {
		return handlermodel.NoteStats{}, fmt.Errorf("count user notes: %w", err)
	}

	stats.TagsCount, err = s.tagCollection.CountDocuments(ctx, tagScopeFilter(scope))
	if err != nil {
		return handlermodel.NoteStats{}, fmt.Errorf("count user tags: %w", err)
	}

	return stats, nil
}

func (s *Storage) Update(noteUUID string, scope storage.Scope, note handlermodel.Note, opts storage.UpdateOptions) (err error) {
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
	delete(updateBody, "user_uuid")
	delete(updateBody, "workspace_id")
	delete(updateBody, "author_user_uuid")
	delete(updateBody, "created_date")
	delete(updateBody, "updated_at")

	if !opts.Body {
		delete(updateBody, "body")
		delete(updateBody, "short_body")
	} else {
		updateBody["body"] = note.Body
		updateBody["short_body"] = note.ShortBody
	}
	if !opts.Header {
		delete(updateBody, "header")
	} else {
		updateBody["header"] = note.Header
	}
	if !opts.Category {
		delete(updateBody, "category_uuid")
	} else {
		updateBody["category_uuid"] = note.CategoryUuid
	}
	if !opts.Tags {
		delete(updateBody, "tags")
	} else {
		updateBody["tags"] = note.Tags
	}
	if !opts.Event {
		delete(updateBody, "event")
	}
	updateBody["updated_at"] = note.UpdatedAt

	update := bson.M{"$set": updateBody}
	if opts.Event {
		if note.Event != nil {
			updateBody["event"] = bson.M{
				"enabled":  note.Event.Enabled,
				"start_at": note.Event.StartAt,
				"end_at":   note.Event.EndAt,
			}
		} else {
			update["$unset"] = bson.M{"event": ""}
			delete(updateBody, "event")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.UpdateResult
	filter := noteAccessFilter(scope)
	filter["_id"] = objectID
	result, err = s.noteCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}

	if result.MatchedCount == 0 {
		return apperror.NotFoundError("note not found")
	}

	return nil
}

func toHandlerEvent(event *mongoNoteEvent) *handlermodel.NoteEvent {
	if event == nil {
		return nil
	}

	return &handlermodel.NoteEvent{
		Enabled: event.Enabled,
		StartAt: event.StartAt,
		EndAt:   event.EndAt,
	}
}

func (s *Storage) Delete(noteUUID string, scope storage.Scope) (err error) {
	objectID, err := primitive.ObjectIDFromHex(noteUUID)
	if err != nil {
		return apperror.BadRequestError("invalid note uuid")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.DeleteResult
	filter := noteAccessFilter(scope)
	filter["_id"] = objectID
	result, err = s.noteCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	if result.DeletedCount == 0 {
		return apperror.NotFoundError("note not found")
	}

	return nil
}

func (s *Storage) DeleteWorkspace(workspaceID string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return apperror.BadRequestError("workspace_id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"workspace_id": workspaceID}

	if _, err := s.noteCollection.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("delete workspace notes: %w", err)
	}

	if _, err := s.tagCollection.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("delete workspace tags: %w", err)
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

func (s *Storage) FindTags(tagUUIDs []string, scope storage.Scope) (tags []tagmodel.Tag, err error) {
	tags = make([]tagmodel.Tag, 0)

	filter := tagScopeFilter(scope)
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
			Uuid:        rawTag.ID.Hex(),
			UserUuid:    rawTag.UserUuid,
			WorkspaceID: rawTag.WorkspaceID,
			Name:        rawTag.Name,
		})
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor tags: %w", err)
	}

	return tags, nil
}

func (s *Storage) CheckTagsExist(tagUUIDs []string, scope storage.Scope) (err error) {
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
	filter := tagScopeFilter(scope)
	filter["_id"] = bson.M{"$in": objectIDs}
	count, err = s.tagCollection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("count tags: %w", err)
	}

	if count != int64(len(objectIDs)) {
		return apperror.NotFoundError("tags not found")
	}

	return nil
}

func (s *Storage) DeleteTag(tagUUID string, scope storage.Scope) (err error) {
	objectID, err := primitive.ObjectIDFromHex(tagUUID)
	if err != nil {
		return apperror.BadRequestError("invalid tag uuid")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.DeleteResult
	filter := tagScopeFilter(scope)
	filter["_id"] = objectID
	result, err = s.tagCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	if result.DeletedCount == 0 {
		return apperror.NotFoundError("tag not found")
	}

	_, err = s.noteCollection.UpdateMany(
		ctx,
		noteScopeWithTagFilter(scope, tagUUID),
		bson.M{"$pull": bson.M{"tags": tagUUID}},
	)
	if err != nil {
		return fmt.Errorf("pull deleted tag from notes: %w", err)
	}

	return nil
}
