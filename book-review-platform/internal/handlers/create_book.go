package handlers

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateBookHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client, ctx context.Context) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	id := primitive.NewObjectID()

	collection := client.Database("book").Collection("books")
	_, err = collection.InsertOne(ctx, bson.M{"_id": id, "title": title, "author": author, "description": description})
	if err != nil {
		http.Error(w, "Error creating book", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}


