package controllers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/api/http/presenters"
	"mime/multipart"
	"net/url"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// Struct untuk menampung hasil dari setiap goroutine
type uploadResult struct {
	Data map[string]interface{}
	Err  error
}

// UploadFiles godoc
//
//	@Summary		Upload files
//	@Description	Upload multiple files to storage
//	@Tags			files
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"File to upload (multiple allowed)"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	presenters.ErrorResponse
//	@Failure		500		{object}	presenters.ErrorResponse
//	@Router			/files/upload [post]
func UploadFiles() fiber.Handler {
	return func(c *fiber.Ctx) error {
		form, err := c.MultipartForm()
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, err.Error())
		}

		// Ambil semua file dengan key "file"
		files := form.File["file"]
		if len(files) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No files provided")
		}

		// Channel untuk menampung hasil upload (buffered channel sesuai jumlah file)
		resultsChan := make(chan uploadResult, len(files))
		var wg sync.WaitGroup

		// Loop setiap file dan jalankan upload secara paralel
		for _, file := range files {
			wg.Add(1)

			// Jalankan Goroutine
			go func(f *multipart.FileHeader) {
				defer wg.Done()

				// Panggil handler
				res, err := handlers.UploadFileHandler(f)

				// Kirim hasil ke channel
				resultsChan <- uploadResult{
					Data: res,
					Err:  err,
				}
			}(file)
		}

		// Tunggu semua proses selesai
		wg.Wait()
		close(resultsChan)

		// Kumpulkan hasilnya
		var successUploads []map[string]interface{}
		var failedUploads []string

		for res := range resultsChan {
			if res.Err != nil {
				failedUploads = append(failedUploads, res.Err.Error())
			} else {
				successUploads = append(successUploads, res.Data)
			}
		}

		// Response handling
		// Jika semua gagal
		if len(successUploads) == 0 && len(failedUploads) > 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "All uploads failed: "+failedUploads[0])
		}

		// Jika ada yang sukses (walaupun sebagian gagal), kembalikan status sukses dengan detail
		responsePayload := fiber.Map{
			"message": "Upload process finished",
			"data":    successUploads,
		}

		// Jika ada error parsial, sertakan infonya di response
		if len(failedUploads) > 0 {
			responsePayload["errors"] = failedUploads
			responsePayload["message"] = "Upload finished with some errors"
		}

		return c.JSON(responsePayload)
	}
}

// GetFileByName godoc
//
//	@Summary		Get file by name
//	@Description	Retrieve a file by its name
//	@Tags			files
//	@Accept			json
//	@Produce		octet-stream
//	@Param			name	path		string	true	"File name"
//	@Success		200		{file}		binary
//	@Failure		400		{object}	presenters.ErrorResponse
//	@Failure		500		{object}	presenters.ErrorResponse
//	@Router			/files/{name} [get]
func GetFileByName() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawName := c.Params("name")
		name, err := url.QueryUnescape(rawName)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		err = handlers.GetFileByNameHandler(c, name)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return nil
	}
}

// DeleteFile godoc
//
//	@Summary		Delete a file
//	@Description	Delete file by name
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Param			name	path	string	true	"File name"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/files/{name} [delete]
func DeleteFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawName := c.Params("name")
		name, err := url.QueryUnescape(rawName)
		fmt.Println("name", name)
		if name == "" {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Object name is required")
		}
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		err = handlers.DeleteFileHandler(name)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "File deleted successfully",
			"object":  name,
		})
	}
}
