package handlers

import (
	"context"
	"io"
	"jk-api/internal/config"
	"mime/multipart"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetFileByNameHandler(c *fiber.Ctx, name string) error {
	ctx := context.Background()

	rc, contentType, err := config.Bucket.GetFile(ctx, name)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "File not found")
	}
	defer rc.Close()

	if contentType != "" {
		c.Set("Content-Type", contentType)
	}

	if _, err := io.Copy(c.Response().BodyWriter(), rc); err != nil {
		return err
	}

	return nil
}

func UploadFileHandler(data *multipart.FileHeader) (map[string]interface{}, error) {
	file, err := data.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ctx := context.Background()
	result, err := config.Bucket.Upload(ctx, file, data)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"filename":    result.Filename,
		"size":        result.Size,
		"object_name": result.ObjectName,
	}, nil
}

func DeleteFileHandler(name string) error {
	ctx := context.Background()
	return config.Bucket.Delete(ctx, name)
}

func GenerateSignedURL(object string, expiry time.Duration) (string, error) {
	return config.Bucket.GenerateSignedURL(object, expiry)
}
