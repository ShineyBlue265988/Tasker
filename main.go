package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gookit/color.v1"
	"log"
	"os"
	"time"
)

var collection *mongo.Collection
var CTX = context.TODO()

type Task struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Text      string             `bson:"text"`
	Completed bool               `bson:"completed"`
}

func createTask(task *Task) error {
	_, err := collection.InsertOne(CTX, task)
	return err
}

func getAll() ([]*Task, error) {
	filter := bson.D{{}}
	return filterTasks(filter)
}

func filterTasks(cursor interface{}) ([]*Task, error) {
	var tasks []*Task
	cur, err := collection.Find(CTX, cursor)
	if err != nil {
		return nil, err
	}
	for cur.Next(CTX) {
		task := &Task{}
		err := cur.Decode(task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if cur.Err() != nil {
		return tasks, cur.Err()
	}
	cur.Close(CTX)
	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}
	return tasks, nil
}

func completeTask(text string) error {
	filter := bson.D{primitive.E{Key: "text", Value: text}}

	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "completed", Value: true},
	}}}

	t := &Task{}
	return collection.FindOneAndUpdate(CTX, filter, update).Decode(t)
}

func getPending() ([]*Task, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: false},
	}

	return filterTasks(filter)
}

func getFinished() ([]*Task, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: true},
	}

	return filterTasks(filter)
}

func deleteTask(text string) error {
	filter := bson.D{primitive.E{Key: "text", Value: text}}

	res, err := collection.DeleteOne(CTX, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("No tasks were deleted")
	}

	return nil
}

func printTasks(tasks []*Task) {
	for i, v := range tasks {
		if v.Completed {
			color.Green.Printf("%d: %s\n", i+1, v.Text)
		} else {
			color.Yellow.Printf("%d: %s\n", i+1, v.Text)
		}
	}
}
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
		Usage: "A simple CLI program to manage your tasks",
		Action: func(c *cli.Context) error {
			tasks, err := getPending()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					fmt.Print("Nothing to see here.\nRun `add 'task'` to add a task")
					return nil
				}

				return err
			}

			printTasks(tasks)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a task to the list",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						return errors.New("Cannot add an empty task")
					}

					task := &Task{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Text:      str,
						Completed: false,
					}

					return createTask(task)
				},
			},
			{
				Name:    "all",
				Aliases: []string{"l"},
				Usage:   "list all tasks",
				Action: func(c *cli.Context) error {
					tasks, err := getAll()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `add 'task'` to add a task")
							return nil
						}

						return err
					}

					printTasks(tasks)
					return nil
				},
			},
			{
				Name:    "done",
				Aliases: []string{"d"},
				Usage:   "complete a task on the list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					return completeTask(text)
				},
			},
			{
				Name:    "finished",
				Aliases: []string{"f"},
				Usage:   "list completed tasks",
				Action: func(c *cli.Context) error {
					tasks, err := getFinished()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `done 'task'` to complete a task")
							return nil
						}

						return err
					}

					printTasks(tasks)
					return nil
				},
			},
			{
				Name:  "rm",
				Usage: "deletes a task on the list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					err := deleteTask(text)
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
