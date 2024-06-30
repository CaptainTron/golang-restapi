package main

import (
	"context"
	"errors"
	"fmt"

	// "encoding/json"

	// "fmt"
	"time"

	// "net/http"
	// "strconv"

	// "github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	db         *mongo.Client
	database   string
	collection string
}

type Server_message struct {
	Message string `json:"message"`
	Status  string `json:"status"`

	// Result    []models.User `json:"result,omitempty"`
	// ResultOne *models.User  `json:"resultone,omitempty"`
}

// Connect to MongoDB
func ConnectToMongoDB(url, database, collection string) (*MongoStore, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))
	if err != nil {
		panic(err)
	}
	return &MongoStore{
		db:         client,
		database:   database,
		collection: collection,
	}, nil
}

func (m *MongoStore) CreateAccount(acc *Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := m.db.Database(m.database).Collection(m.collection)
	// acc.ID= primitive.NewObjectID()
	_, err := coll.InsertOne(ctx, acc)
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoStore) DeleteAccount(id int) error {

	coll := m.db.Database(m.database).Collection(m.collection)
	// idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"id", id}}

	DeletedCount, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if DeletedCount.DeletedCount != 1 {
		return fmt.Errorf("No Account found with this Id: %d", id)
	}
	return nil
}

func (m *MongoStore) UpdateAccount(acc *Account) error {

	coll := m.db.Database(m.database).Collection(m.collection)
	// idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"id", acc.ID}}

	update := bson.D{{"$set", bson.D{{"firstname", acc.FirstName}, {"lastname", acc.LastName}, {"balance", acc.Balance}}}}
	UpdateResult, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if UpdateResult.MatchedCount != 1 {
		return fmt.Errorf("no user found with id: %d", acc.ID)
	}
	return nil
}
func (m *MongoStore) GetAccounts() ([]*Account, error) {
	coll := m.db.Database(m.database).Collection(m.collection)
	filter := bson.D{{}}

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	acc := []*Account{}
	if err = cursor.All(context.TODO(), &acc); err != nil {
		panic(err)
	}

	return acc, nil
}
func (m *MongoStore) GetAccountByID(id int) (*Account, error) {
	coll := m.db.Database(m.database).Collection(m.collection)
	// idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"id", id}}

	account := Account{}
	err := coll.FindOne(context.TODO(), filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		panic(err)
	}
	return &account, nil
}

func (m *MongoStore) TransferAmount(fromID, toID, amount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := m.db.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if abortErr := session.AbortTransaction(ctx); abortErr != nil {
				fmt.Println("Error aborting transaction:", abortErr)
			}
		}
	}()

	coll := m.db.Database(m.database).Collection(m.collection)

	// Update sender balance (with check for sufficient funds)
	updateResult, err := coll.UpdateOne(ctx, bson.M{"id": fromID, "balance": bson.M{"$gte": amount}}, bson.M{"$inc": bson.M{"balance": -amount}})
	if err != nil {
		return err
	}

	if updateResult.ModifiedCount == 0 {
		return errors.New("Insufficient funds")
	}

	// Update receiver balance
	_, err = coll.UpdateOne(ctx, bson.M{"id": toID}, bson.M{"$inc": bson.M{"balance": amount}})
	if err != nil {
		return err
	}

	err = session.CommitTransaction(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Transfer successful!")
	return nil
}
