package handlers

import (
	"context"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func editBookHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client, ctx context.Context) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}
	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	collection := client.Database("book").Collection("books")
	var book Book
	err = collection.FindOne(ctx, bson.M{"title": title}).Decode(&book)
	if err != nil {
		log.Printf("Error retrieving book details: %v\n", err)
		http.Error(w, "Error retrieving book details", http.StatusInternalServerError)
		return
	}

	book.Author = author
	book.Description = description

	_, err = collection.ReplaceOne(ctx, bson.M{"title": title}, book)
	if err != nil {
		log.Printf("Error updating book review: %v\n", err)
		http.Error(w, "Error updating book review", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
