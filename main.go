package main

import (
	"context"
	// "fmt"
	// "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// "github.com/julienschmidt/httprouter"

	"example/learn/controllers"
	// "net/http"
)

func main() {
	// r := httprouter.New()
	session := ConnectToMongo("mongodb://127.0.0.1:27017")
	defer session.Disconnect(context.TODO())
	uc := controllers.NewController(session, "sample", "sample")

	uc.GetAllUser()
	// uc.PostUser()
	// uc.UpdateUser("667aaa2420ec1d3d1425cbfe")
	// uc.GetUser("667aa3b4cb4eda3594be89d4")
	// uc.DeleteUser("000000000000000000000000")

	// fmt.Println("Done!")

	// r.GET("/user/:id", uc.GetUser)
	// r.DELETE("/user/:id", uc.GetAllUser)
	// r.POST("/user", uc.PostUser)
	// r.DELETE("/user/:id", uc.DeleteUser)

	// fmt.Println("Server is Listening to port 9000...")
	// http.ListenAndServe("localhost:9000", r)
}

func ConnectToMongo(uri string) *mongo.Client {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	return client
}
