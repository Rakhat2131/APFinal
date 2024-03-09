package handlers

import (
	"context"
	"log"
	"net/http"
	"text/template"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Book struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Author      string             `bson:"author"`
	Description string             `bson:"description"`
}

func IndexPageHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client, ctx context.Context) {
	collection := client.Database("book").Collection("books")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error retrieving book reviews: %v\n", err)
		http.Error(w, "Error retrieving book reviews", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var books []Book
	for cursor.Next(ctx) {
		var book Book
		if err := cursor.Decode(&book); err != nil {
			log.Printf("Error decoding book review: %v\n", err)
			http.Error(w, "Error decoding book review", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	tmpl, err := template.ParseFiles("./web/templates/index.html")
	if err != nil {
		log.Printf("Error parsing HTML template: %v\n", err)
		http.Error(w, "Error parsing HTML template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Books []Book
	}{
		Books: books,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing HTML template: %v\n", err)
		http.Error(w, "Error executing HTML template", http.StatusInternalServerError)
		return
	}
}
