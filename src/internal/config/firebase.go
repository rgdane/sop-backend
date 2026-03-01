package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
	"gorm.io/datatypes"
)

var (
	fcmApp     *firebase.App
	fcmOnce    sync.Once
	fcmMsgOnce sync.Once

	authClient *auth.Client
	msgClient  *messaging.Client
)

func FirebaseApp() error {
	opt := option.WithCredentialsFile(filepath.Join("internal", "config", "creds", "gcp_firebase.json"))
	var err error
	fcmApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	return nil
}

func GetFirebaseCreds() (datatypes.JSON, error) {
	path := filepath.Join("internal", "config", "creds", "gcp_firebase.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(data), nil
}

func GetFirebaseAuth() *auth.Client {
	if fcmApp == nil {
		Logger.Error("❌ Firebase app is not initialized")
		return nil
	}
	fcmOnce.Do(func() {
		var err error
		authClient, err = fcmApp.Auth(context.Background())
		if err != nil {
			Logger.Errorf("❌ Failed to get Firebase Auth client: %v", err)
		}
	})
	return authClient
}

func GetFirebaseMessaging() *messaging.Client {
	if fcmApp == nil {
		Logger.Error("❌ Firebase app is not initialized")
		return nil
	}
	fcmMsgOnce.Do(func() {
		var err error
		msgClient, err = fcmApp.Messaging(context.Background())
		if err != nil {
			Logger.Errorf("❌ Failed to get Firebase Messaging client: %v", err)
		}
	})
	return msgClient
}
