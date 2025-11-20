package main

import (
	"fmt"
	"log"
	"os"

	"secure-fm/auth"
	"secure-fm/config"
	"secure-fm/db"
	"secure-fm/fs"
	"secure-fm/utils"
)

var currentUser *db.User

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.LoadConfig()

	db.InitDB(cfg)
	fs.InitFS(cfg)

	fmt.Println("Welcome to Secure File Manager")

	for {
		if currentUser == nil {
			authMenu()
		} else {
			mainMenu()
		}
	}
}

func authMenu() {
	fmt.Println("\n--- Auth Menu ---")
	fmt.Println("1. Login")
	fmt.Println("2. Register")
	fmt.Println("3. Exit")

	choice := utils.ReadLine("Select option: ")

	switch choice {
	case "1":
		login()
	case "2":
		register()
	case "3":
		os.Exit(0)
	default:
		fmt.Println("Invalid option")
	}
}

func login() {
	username := utils.ReadLine("Username: ")
	password := utils.ReadLine("Password: ")

	user, err := db.GetUserByUsername(username)
	if err != nil {
		fmt.Println("Error fetching user:", err)
		return
	}
	if user == nil {
		fmt.Println("Invalid username or password")
		return
	}

	if auth.CheckPasswordHash(password, user.PasswordHash) {
		currentUser = user
		fmt.Println("Login successful!")
	} else {
		fmt.Println("Invalid username or password")
	}
}

func register() {
	username := utils.ReadLine("Username: ")
	password := utils.ReadLine("Password: ")

	if len(password) < 8 {
		fmt.Println("Password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}

	err = db.CreateUser(username, hash)
	if err != nil {
		fmt.Println("Error creating user (username might be taken):", err)
		return
	}
	fmt.Println("Registration successful! Please login.")
}

func mainMenu() {
	fmt.Println("\n--- Main Menu ---")
	fmt.Println("Logged in as:", currentUser.Username)
	fmt.Println("1. List Drives / System Info")
	fmt.Println("2. List Directory")
	fmt.Println("3. Read File")
	fmt.Println("4. Write File")
	fmt.Println("5. Delete File")
	fmt.Println("6. Copy File")
	fmt.Println("7. Move File")
	fmt.Println("8. Read JSON")
	fmt.Println("9. Write JSON")
	fmt.Println("10. Read XML")
	fmt.Println("11. Write XML")
	fmt.Println("12. Create ZIP")
	fmt.Println("13. Extract ZIP")
	fmt.Println("14. Logout")

	choice := utils.ReadLine("Select option: ")

	switch choice {
	case "1":
		drives := fs.ListDrives()
		fmt.Println("Drives/Mounts:", drives)
		db.LogOperation("list_drives", 0, currentUser.ID)
	case "2":
		path := utils.ReadLine("Path (relative to sandbox): ")
		files, err := fs.ListDirectory(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			for _, f := range files {
				fmt.Printf("%s \t %d bytes \t %s\n", f.Name(), f.Size(), f.Mode())
			}
		}
		db.LogOperation("list_dir", 0, currentUser.ID)
	case "3":
		path := utils.ReadLine("Path: ")
		content, err := fs.ReadFile(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Content:\n", content)
		}
		db.LogOperation("read_file", 0, currentUser.ID)
	case "4":
		path := utils.ReadLine("Path: ")
		content := utils.ReadLine("Content: ")
		err := fs.WriteFile(path, content)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("File written successfully")
			// Log metadata
			id, _ := db.CreateFileMetadata(path, int64(len(content)), path, currentUser.ID)
			db.LogOperation("write_file", id, currentUser.ID)
		}
	case "5":
		path := utils.ReadLine("Path: ")
		err := fs.DeleteFile(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("File deleted")
			db.LogOperation("delete_file", 0, currentUser.ID)
		}
	case "6":
		src := utils.ReadLine("Source Path: ")
		dst := utils.ReadLine("Dest Path: ")
		err := fs.CopyFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("File copied")
			db.LogOperation("copy_file", 0, currentUser.ID)
		}
	case "7":
		src := utils.ReadLine("Source Path: ")
		dst := utils.ReadLine("Dest Path: ")
		err := fs.MoveFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("File moved")
			db.LogOperation("move_file", 0, currentUser.ID)
		}
	case "8":
		path := utils.ReadLine("Path: ")
		data, err := fs.ReadJSON(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Data: %+v\n", data)
		}
		db.LogOperation("read_json", 0, currentUser.ID)
	case "9":
		path := utils.ReadLine("Path: ")
		key := utils.ReadLine("Key: ")
		val := utils.ReadLine("Value: ")
		data := map[string]string{key: val}
		err := fs.WriteJSON(path, data)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("JSON written")
			db.LogOperation("write_json", 0, currentUser.ID)
		}
	case "10":
		path := utils.ReadLine("Path: ")
		data, err := fs.ReadXML(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Data: %+v\n", data)
		}
		db.LogOperation("read_xml", 0, currentUser.ID)
	case "11":
		path := utils.ReadLine("Path: ")
		content := utils.ReadLine("Content: ")
		data := &fs.XMLData{Content: content}
		err := fs.WriteXML(path, data)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("XML written")
			db.LogOperation("write_xml", 0, currentUser.ID)
		}
	case "12":
		src := utils.ReadLine("Source Dir/File: ")
		dst := utils.ReadLine("Dest Zip Path: ")
		err := fs.CreateZip(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Zip created")
			db.LogOperation("create_zip", 0, currentUser.ID)
		}
	case "13":
		src := utils.ReadLine("Zip Path: ")
		dst := utils.ReadLine("Dest Dir: ")
		err := fs.Unzip(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Zip extracted")
			db.LogOperation("extract_zip", 0, currentUser.ID)
		}
	case "14":
		currentUser = nil
		fmt.Println("Logged out")
	default:
		fmt.Println("Invalid option")
	}
}
