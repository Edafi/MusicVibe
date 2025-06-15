package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Edafi/MusicVibe/routes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI("mongodb://username:password@localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client.Database("audiostreaming"), nil
}

func main() {
	db, err := sql.Open("mysql", "golang:QXmREMXPd321Edafi@tcp(localhost:3306)/audiostream")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := "mongodb://golang:QXmREMXPd321Edafi@localhost:27017/audiostreaming"

	clientOptions := options.Client().ApplyURI(uri)

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	defer mongoClient.Disconnect(ctx)

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
	}

	mongoDatabase := mongoClient.Database("audiostreaming")

	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	handler := routes.SetupRoutes(db, mongoDatabase, minioClient)
	log.Println("Server running on :8080")
	log.Println(http.ListenAndServe(":8080", handler))
}
