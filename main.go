package make

import (
	"context"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

var collection *mongo.Collection
var CTX = context.TODO()

func init() {
	client, err := mongo.Connect(CTX, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(CTX, nil)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("tasker").Collection("task")

}

func main() {
	app := &cli.App{
		Name:  "tasker",
		Usage: "Simple CLI program to manage your task.",
		Commands: []*cli.Command{
			{},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
