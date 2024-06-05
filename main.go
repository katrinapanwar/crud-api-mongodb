package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Define the Car structure
type Car struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	CarName string             `bson:"carName"`
	Color   string             `bson:"color"`
}

// Define the CarManager interface
type CarManager interface {
	Insert(Car) error
	GetAll() ([]Car, error)
	DeleteData(primitive.ObjectID) error
	UpdateData(Car) error
}

// Implement the CarManager struct
type carManager struct {
	connection *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

var Mgr CarManager

func connectDb() {
	uri := "localhost:27017"
	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("%s%s", "mongodb://", uri)))
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected!!!")
	Mgr = &carManager{connection: client, ctx: ctx, cancel: cancel}
}

func close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func init() {
	connectDb()
}

func main() {
	// Insert record to MongoDB
	car := Car{CarName: "Tesla Model M", Color: "White"}
	err := Mgr.Insert(car)
	fmt.Println(err)

	// Get all records in the DB
	data, err := Mgr.GetAll()
	fmt.Println(data, err)

	// Delete record from DB
	id := "665ff328c8cc0f4aca6dff27"
	// objectId, err := primitive.ObjectIDFromHex(id)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// err = Mgr.DeleteData(objectId)
	// fmt.Println(err)

	// Update
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	car.ID = objectId
	car.CarName = "Virtus"
	car.Color = "Navy Blue"
	err = Mgr.UpdateData(car)
	fmt.Println(err)
}

func (mgr *carManager) Insert(data Car) error {
	carCollection := mgr.connection.Database("Katrina").Collection("testdb")
	result, err := carCollection.InsertOne(context.TODO(), data)
	fmt.Println(result.InsertedID)
	return err
}

func (mgr *carManager) GetAll() (data []Car, err error) {
	carCollection := mgr.connection.Database("Katrina").Collection("testdb")

	// Pass these options to the Find method
	findOptions := options.Find()

	cur, err := carCollection.Find(context.TODO(), bson.M{}, findOptions)
	for cur.Next(context.TODO()) {
		var d Car
		err := cur.Decode(&d)
		if err != nil {
			log.Fatal(err)
		}
		data = append(data, d)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	// Close the cursor once finished
	cur.Close(context.TODO())

	return data, nil
}

func (mgr *carManager) DeleteData(id primitive.ObjectID) error {
	carCollection := mgr.connection.Database("Katrina").Collection("testdb")

	filter := bson.D{{"_id", id}}
	_, err := carCollection.DeleteOne(context.TODO(), filter)
	return err
}

func (mgr *carManager) UpdateData(data Car) error {
	carCollection := mgr.connection.Database("Katrina").Collection("testdb")

	filter := bson.D{{"_id", data.ID}}
	update := bson.D{{"$set", data}}

	_, err := carCollection.UpdateOne(context.TODO(), filter, update)

	return err
}
