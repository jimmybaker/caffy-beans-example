package db

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

func ProvideDB() *firestore.Client {
	projectID := "astute-charter-282823"

	client, err := firestore.NewClient(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return client
}

var Options = ProvideDB
