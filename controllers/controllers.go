package controllers

import (
	"context"
	"encoding/json"
	"example/learn/models"
	"fmt"
	// "net/http"

	// "github.com/julienschmidt/httprouter"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewController(s *mongo.Client, database, collection string) *UserController {
	return &UserController{
		client:     s,
		database:   database,
		collection: collection,
	}
}

func (uc UserController) GetAllUser() {
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	filter := bson.D{{}}

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	var result []models.User
	if err = cursor.All(context.TODO(), &result); err != nil {
		panic(err)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", output)
}

func (uc UserController) GetUser(id string) {
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	filter := bson.D{{"Id", id}}

	var result UserController
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No Document found!!")
			return
		}
		panic(err)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", output)

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// fmt.Fprintf(w, "%s\n", uj)
}

func (uc UserController) PostUser() {
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	newUser := models.User{
		Name:   "Vaibhav Yadav",
		Gender: "Male",
		Age:    50234,
	}

	result, err := coll.InsertOne(context.TODO(), newUser)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Document: %s, ID: %s\n", result, result.InsertedID)

	// return result.InsertedID
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// fmt.Fprintf(w, "%s\n", uj)
}

func (uc UserController) UpdateUser(id string) {
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", idNumber}}

	update := bson.D{{"$set", bson.D{{"name", "Shubham Yadav"}, {"gender", "M"}, {"age", 1000}}}}

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Documents updated: %v\n", result.ModifiedCount)
}

func (uc UserController) DeleteUser(id string) {
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", idNumber}}

	result, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Documents deleted: %d\n", result.DeletedCount)

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// fmt.Fprint(w, "Deleted user", oid, "\n")
}
