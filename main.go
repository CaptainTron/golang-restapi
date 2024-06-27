package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/julienschmidt/httprouter"
	"example/learn/controllers"
	"net/http"
)

func main() {
	r := httprouter.New()

	session := ConnectToMongo("mongodb://127.0.0.1:27017")
	defer session.Disconnect(context.TODO())
	uc := controllers.NewController(session, "sample", "sample")



	r.GET("/users", uc.GetAllUser)
	r.GET("/user/:id", uc.GetUser)
	r.POST("/createuser", uc.PostUser)
	r.PATCH("/updateuser/:id", uc.UpdateUser)
	r.DELETE("/user/:id", uc.DeleteUser)

	fmt.Println("Server is Listening to port 9000...")
	http.ListenAndServe("localhost:9000", r)
}

func ConnectToMongo(uri string) *mongo.Client {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	return client
}
