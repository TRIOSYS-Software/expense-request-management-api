package configs

import (
	"context"
	"log"
	"os"
	"path/filepath"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func (c *Config) SetupFirebase() error {
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		credPath = filepath.Join(".", "fcm_credentials.json")
	}
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		log.Printf("Firebase credentials file does not exist at path: %s\n", credPath)
		return err
	}
	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Printf("error initializing Firebase app: %v\n", err)
		return err
	}
	log.Printf("Firebase app initialized: %+v\n", app)
	log.Printf("Option data: %+v\n", opt)
	c.FirebaseApp = app
	return nil
}
