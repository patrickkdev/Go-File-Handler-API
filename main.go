package main

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fileUtils "github.com/patrickkdev/go-file-handler/utils"
)

// REST API that handles folders and files in the folder storage. Also handle file download and upload.
func main() {
	app := fiber.New()
	app.Server().MaxRequestBodySize = 1073741824

	// Logger Middleware
	app.Use(logger.New(logger.Config{
		// Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		Format: "${method} ${path} - ${status}\n",
	}))

	app.Get("/*/download", handleDownload)
	app.Get("/*", handleGetFolderStructure)
	
	app.Put("/*", handleCreateFolder)
	app.Post("/*", handleUpload)
	app.Patch("/*", handleRename)

	app.Delete("/*", handleDelete)

	app.Listen(":9990")
}

func formatPath(path string) string {
	return fileUtils.ReplaceSpecialChars(strings.ReplaceAll(fileUtils.ROOT + "/" + path, "%20", " "))
}

func handleUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(500).SendString("Error uploading file: " + err.Error())
	}

	targetPath := formatPath(c.Params("*"))

	// Customize your storage path as needed
	storagePath := targetPath + "/" + file.Filename

	// Create nested folders if they don't exist
	if err := fileUtils.MkdirAll(strings.TrimSuffix(storagePath, file.Filename)); err != nil {
		return c.Status(500).SendString("Error creating nested folders: " + err.Error())
	}

	// Save the file
	if err := c.SaveFile(file, storagePath); err != nil {
		return c.Status(500).SendString("Error saving file: " + err.Error())
	}

	return c.SendString("File uploaded successfully")
}

func handleDownload(c *fiber.Ctx) error {
    filePath := formatPath(c.Params("*"))

    isFile, err := fileUtils.FileIsFile(filePath)
    if err != nil {
        return c.Status(http.StatusInternalServerError).SendString("Error checking file: " + err.Error())
    }

    if !isFile {
        return c.Status(http.StatusNotFound).SendString("File not found")
    }

    // Get the filename from the filePath
    filename := filepath.Base(filePath)

    // Set the Content-Disposition header to specify the filename
    c.Set("Content-Disposition", "attachment; filename="+filename)

    // Send the file as a response
    return c.SendFile(filePath)
}

func handleCreateFolder(c *fiber.Ctx) error {
	folderPath := formatPath(c.Params("*"))

	isDir, _ := fileUtils.DirIsDir(folderPath)
	
	if isDir {
		return c.Status(200).SendString("Folder already exists")
	}

	if err := fileUtils.MkdirAll(folderPath); err != nil {
		return c.Status(500).SendString("Error creating folder: " + err.Error())
	}

	return c.Status(200).SendString("Folder created successfully")
}

func handleDelete(c *fiber.Ctx) error {
	// Check auth before running destructive action
	password := c.FormValue("password")
	useForce := c.FormValue("force") == "true"

	if password != "Junio020499" {
		return c.Status(401).SendString("Unauthorized: Wrong password")
	}

	folderPath := formatPath(c.Params("*"))

	isDir, _ := fileUtils.DirIsDir(folderPath)

	if err := fileUtils.Remove(folderPath, useForce); err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "not empty") {
			errorMessage += " Delete all files inside the folder before deleting it or pass 'force=true' as a query parameter."
		}

		return c.Status(500).SendString("Error deleting folder: " + errorMessage)
	}

	if (isDir) {
		return c.Status(200).SendString("Folder deleted successfully")
	} 

	return c.Status(200).SendString("File deleted successfully")
}

func handleRename(c *fiber.Ctx) error {
	targetPath := formatPath(c.Params("*"))
	newPath := formatPath(c.FormValue("newPath"))

	err := fileUtils.Rename(targetPath, newPath)

	if err != nil {
		return c.Status(500).SendString("Error renaming folder: " + err.Error())
	}

	return c.Status(200).SendString("Folder renamed successfully")
}

func handleGetFolderStructure(c *fiber.Ctx) error {
	targetPath := formatPath(c.Params("*"))

	folderStructure, err := fileUtils.GetFolderStructure(targetPath)
	if err != nil {
		return c.Status(500).SendString("Error getting folder structure: " + err.Error())
	}

	// You can serialize folderStructure to JSON and send it as the response
	return c.Status(200).JSON(folderStructure)
}
