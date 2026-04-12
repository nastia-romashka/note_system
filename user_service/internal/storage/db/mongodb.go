package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"user_service/internal/apperror"
	"user_service/internal/config"
	handlermodel "user_service/internal/handlers/users"
	"user_service/pkg/logging"
)

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type mongoUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string             `bson:"username,omitempty"`
	Email        string             `bson:"email,omitempty"`
	PasswordHash string             `bson:"password_hash,omitempty"`
	CreatedDate  int64              `bson:"created_date,omitempty"`
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
		client:     client,
		collection: client.Database(cfg.Mongo.Database).Collection(cfg.Mongo.Collection),
	}

	_, err = storage.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		logger.Error("failed to create user indexes", "error", err)
		return nil, fmt.Errorf("create user indexes: %w", err)
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

func (s *Storage) Create(user handlermodel.User) (userUUID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *mongo.InsertOneResult
	result, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", apperror.BadRequestError("user already exists")
		}
		return "", fmt.Errorf("insert user: %w", err)
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("inserted id is not object id")
	}

	userUUID = objectID.Hex()
	return userUUID, nil
}

func (s *Storage) FindOne(userUUID string) (user handlermodel.User, err error) {
	var objectID primitive.ObjectID
	objectID, err = primitive.ObjectIDFromHex(userUUID)
	if err != nil {
		return handlermodel.User{}, apperror.BadRequestError("invalid user uuid")
	}

	user, err = s.findOne(bson.M{"_id": objectID})
	if err != nil {
		return handlermodel.User{}, err
	}

	user.Uuid = userUUID
	return user, nil
}

func (s *Storage) FindByUsername(username string) (user handlermodel.User, err error) {
	return s.findOne(bson.M{"username": username})
}

func (s *Storage) findOne(filter bson.M) (user handlermodel.User, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := s.collection.FindOne(ctx, filter)
	if err = result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return handlermodel.User{}, apperror.NotFoundError("user not found")
		}
		return handlermodel.User{}, fmt.Errorf("find user: %w", err)
	}

	var rawUser mongoUser
	if err = result.Decode(&rawUser); err != nil {
		return handlermodel.User{}, fmt.Errorf("decode user: %w", err)
	}

	user = handlermodel.User{
		Uuid:         rawUser.ID.Hex(),
		Username:     rawUser.Username,
		Email:        rawUser.Email,
		PasswordHash: rawUser.PasswordHash,
		CreatedDate:  rawUser.CreatedDate,
	}

	return user, nil
}
