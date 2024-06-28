package controllers

import (
	"context"
	"encoding/json"
	"example/learn/models"
	"fmt"
	"time"

	"net/http"
	"strconv"


	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	client     *mongo.Client
	database   string
	collection string
}

type Server_message struct {
	Message string `json: "message"`
	Status  string `json: "status"`

	Result    []models.User `json:"result,omitempty"`
	ResultOne *models.User  `json:"resultone,omitempty"`
}

func NewController(s *mongo.Client, database, collection string) *UserController {
	return &UserController{
		client:     s,
		database:   database,
		collection: collection,
	}
}

// Get all the user
func (uc UserController) GetAllUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	servermessage := Server_message{"Total Number of Records " + strconv.Itoa(len(result)), "Success", result, nil}

	output, err := json.MarshalIndent(servermessage, "", "    ")
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", output)
}

// Fetch Single user
func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	id := p.ByName("id")
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", idNumber}}

	var result models.User
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			document_err := Server_message{string(id) + " Not found", "Not Successful", nil, nil}
			output, _ := json.MarshalIndent(document_err, "", "    ")
			fmt.Fprintf(w, "%s\n", output)
			return
		}
		panic(err)
	}
	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", output)
}

// Create User
func (uc UserController) PostUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coll := uc.client.Database(uc.database).Collection(uc.collection)
	var newUser models.User


	err := json.NewDecoder(r.Body).Decode(&newUser)
	handle_error(err)


	newUser.Id = primitive.NewObjectID()
	_, err = coll.InsertOne(ctx, newUser)
	handle_error(err)

	message := Server_message{"User Created Successfully", "User Created", nil, nil}
	output, err := json.MarshalIndent(message, "", "    ")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\n", output)
}

// UpdateUser
func (uc UserController) UpdateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	id := p.ByName("id")
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", idNumber}}

	var update_user models.User
	err := json.NewDecoder(r.Body).Decode(&update_user)
	handle_error(err)

	update := bson.D{{"$set", bson.D{{"name", update_user.Name}, {"gender", update_user.Gender}, {"age", update_user.Age}}}}
	UpdateResult, err := coll.UpdateOne(context.TODO(), filter, update)
	if UpdateResult.MatchedCount != 1 {
		w.WriteHeader(http.StatusNotFound)

		message := Server_message{"No User Found with given Id", "user does not exist", nil, nil}
		output, err := json.MarshalIndent(message, "", "    ")
		handle_error(err)

		fmt.Fprintf(w, "%s", output)
		return
	}
	handle_error(err)

	message := Server_message{"Operation Success", "User Updated", nil, nil}

	output, err := json.MarshalIndent(message, "", "    ")
	handle_error(err)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", output)
}

// DeleteUser
func (uc UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	id := p.ByName("id")
	coll := uc.client.Database(uc.database).Collection(uc.collection)
	idNumber, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", idNumber}}

	DeletedCount, err := coll.DeleteOne(context.TODO(), filter)
	handle_error(err)

	if DeletedCount.DeletedCount != 1 {
		w.WriteHeader(http.StatusNotFound)
		message := Server_message{"No User found with given Id", "Not found", nil, nil}
		output, _ := json.MarshalIndent(message, "", "    ")
		fmt.Fprintf(w, "%s", output)
		return
	}

	message := Server_message{"User Successfully deleted", "Success", nil, nil}
	output, err := json.MarshalIndent(message, "", "    ")
	handle_error(err)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", output)
}

// This One will handle the error
func handle_error(err error) {
	if err != nil {
		panic(err)
		return
	}
}
