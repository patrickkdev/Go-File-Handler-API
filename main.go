package main

import (
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	fileUtils "github.com/patrickkdev/go-file-handler/utils"
)

const ROOT = "storage-root"

// REST API that handles folders and files in the folder storage. Also handle file download and upload.
func main() {
	app := fiber.New()
	app.Server().MaxRequestBodySize = 1073741824

	app.Get("/", func(c *fiber.Ctx) error {
		c.Status(200).SendString("Hello, World!")

		return nil
	})

	app.Get("/download/*", handleDownload)
	app.Get("/folder-structure/*", handleGetFolderStructure)
	
	app.Post("/upload/*", handleUpload)
	app.Post("/create-folder/*", handleCreateFolder)
	
	app.Delete("/*", handleDelete)

	app.Listen(":3011")
}

func handleUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		c.Status(500).SendString("Error uploading file")
		return err
	}

	targetPath := strings.ReplaceAll(ROOT + "/" + c.Params("*"), "%20", " ")

	// Customize your storage path as needed
	storagePath := targetPath + "/" + file.Filename

	// Create nested folders if they don't exist
	if err := fileUtils.MkdirAll(strings.TrimSuffix(storagePath, file.Filename)); err != nil {
		c.Status(500).SendString("Error creating nested folders")
		return err
	}

	// Save the file
	if err := c.SaveFile(file, storagePath); err != nil {
		c.Status(500).SendString("Error saving file")
		return err
	}

	c.SendString("File uploaded successfully")

	return nil
}

func handleDownload(c *fiber.Ctx) error {
	filePath := strings.ReplaceAll(ROOT + "/" + c.Params("*"), "%20", " ")

	isFile, err := fileUtils.FileIsFile(filePath)
	if err != nil {
		c.Status(500).SendString("Error checking file")
		return err
	}

	if !isFile {
		c.Status(404).SendString("File not found")
		return new(os.PathError)
	}

	c.SendFile(filePath)

	return nil
}

func handleCreateFolder(c *fiber.Ctx) error {
	folderPath := strings.ReplaceAll(ROOT + "/" + c.Params("*"), "%20", " ")

	isDir, _ := fileUtils.DirIsDir(folderPath)
	if isDir {
		c.Status(200).SendString("Folder already exists")
		return nil
	}

	if err := fileUtils.MkdirAll(folderPath); err != nil {
		c.Status(500).SendString("Error creating folder")
		return err
	}

	c.SendString("Folder created successfully")

	return nil
}

func handleDelete(c *fiber.Ctx) error {
	// Check auth before running destructive action
	password := c.FormValue("password")

	if password != "Junio020499" {
		c.Status(401).SendString("Unauthorized")
		return errors.New("Unauthorized")
	}

	folderPath := strings.ReplaceAll(ROOT + "/" + c.Params("*"), "%20", " ")

	isDir, _ := fileUtils.DirIsDir(folderPath)

	if err := os.Remove(folderPath); err != nil {
		c.Status(500).SendString("Error deleting folder")
		return err
	}

	if (isDir) {
		c.SendString("Folder deleted successfully")
	} else {
		c.SendString("File deleted successfully")
	}

	return nil
}

func handleGetFolderStructure(c *fiber.Ctx) error {
	targetPath := strings.ReplaceAll(ROOT + "/" + c.Params("*"), "%20", " ")

	folderStructure, err := fileUtils.GetFolderStructure(targetPath)
	if err != nil {
		c.Status(500).SendString("Error getting folder structure")
		return err
	}

	// You can serialize folderStructure to JSON and send it as the response
	c.JSON(folderStructure)

	return nil
}