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
var currentDir string = "." // текущая директория относительно sandbox

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

// resolveCwd объединяет текущую директорию с введённым путём
func resolveCwd(inputPath string) string {
	if inputPath == "" || inputPath == "." {
		return currentDir
	}
	if inputPath == "/" {
		return "."
	}
	if currentDir == "." {
		return inputPath
	}
	return currentDir + "/" + inputPath
}

// changeDirectory меняет текущую директорию
func changeDirectory(newDir string) error {
	var targetDir string

	switch newDir {
	case "", ".":
		return nil // остаёмся на месте
	case "/":
		targetDir = "."
	case "..":
		if currentDir == "." {
			return nil // уже в корне
		}
		// Получаем родительскую директорию
		lastSlash := -1
		for i := len(currentDir) - 1; i >= 0; i-- {
			if currentDir[i] == '/' {
				lastSlash = i
				break
			}
		}
		if lastSlash == -1 {
			targetDir = "."
		} else {
			targetDir = currentDir[:lastSlash]
		}
	default:
		if currentDir == "." {
			targetDir = newDir
		} else {
			targetDir = currentDir + "/" + newDir
		}
	}

	// Проверяем что директория существует и безопасна
	_, err := fs.ListDirectory(targetDir)
	if err != nil {
		return err
	}

	currentDir = targetDir
	return nil
}

func mainMenu() {
	fmt.Println("\n--- Main Menu ---")
	fmt.Printf("👤 User: %s\n", currentUser.Username)
	fmt.Printf("📂 Current directory: /%s\n", currentDir)
	fmt.Println("────────────────────────────────")
	fmt.Println("0. Change Directory (cd)")
	fmt.Println("1. List Drives / System Info")
	fmt.Println("2. List Directory")
	fmt.Println("3. Create Directory")
	fmt.Println("4. Read File")
	fmt.Println("5. Write File")
	fmt.Println("6. Delete File")
	fmt.Println("7. Copy File")
	fmt.Println("8. Move File")
	fmt.Println("9. Read JSON")
	fmt.Println("10. Write JSON")
	fmt.Println("11. Read XML")
	fmt.Println("12. Write XML")
	fmt.Println("13. Create ZIP")
	fmt.Println("14. Extract ZIP")
	fmt.Println("15. Logout")

	choice := utils.ReadLine("Select option: ")

	switch choice {
	case "0":
		fmt.Println("\n📂 Смена текущей директории")
		fmt.Println("   Подсказка: введите '..' для перехода наверх")
		fmt.Println("   Пример: docs, .., subdir, / (корень sandbox)")
		newDir := utils.ReadLine("New directory: ")
		if err := changeDirectory(newDir); err != nil {
			fmt.Println("❌ Error:", err)
		} else {
			fmt.Printf("✅ Перешли в: /%s\n", currentDir)
		}

	case "1":
		fmt.Println("\n📀 Информация о дисках/файловой системе")
		drives := fs.ListDrives()
		fmt.Println("Доступные разделы:", drives)

		// Отображаем информацию о корневом разделе
		diskInfo, err := fs.GetDiskInfo("/")
		if err == nil {
			fmt.Printf("\n📊 Раздел: %s\n", diskInfo.Name)
			fmt.Printf("   Всего:     %.2f GB\n", float64(diskInfo.TotalSize)/(1024*1024*1024))
			fmt.Printf("   Свободно:  %.2f GB\n", float64(diskInfo.FreeSpace)/(1024*1024*1024))
			fmt.Printf("   Занято:    %.2f GB (%.1f%%)\n", float64(diskInfo.UsedSpace)/(1024*1024*1024), diskInfo.UsedPercent)
		} else {
			fmt.Println("   Не удалось получить информацию о диске:", err)
		}
		db.LogOperation("list_drives", 0, currentUser.ID)

	case "2":
		fmt.Println("\n📁 Просмотр содержимого текущей директории")
		fmt.Println("   (или введите путь для просмотра другой папки)")
		fmt.Println("   Пример: . (текущая), docs, subdir/nested")
		inputPath := utils.ReadLine("Path [. = current]: ")
		path := resolveCwd(inputPath)
		files, err := fs.ListDirectory(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("\n📂 Содержимое /%s:\n", path)
			if len(files) == 0 {
				fmt.Println("   (директория пуста)")
			}
			for _, f := range files {
				fileType := "📄"
				if f.IsDir() {
					fileType := "📁"
					fmt.Printf("   %s %s/\n", fileType, f.Name())
				} else {
					fmt.Printf("   %s %s \t %d bytes\n", fileType, f.Name(), f.Size())
				}
			}
		}
		db.LogOperation("list_dir", 0, currentUser.ID)

	case "3":
		fmt.Println("\n📂 Создание новой директории")
		fmt.Println("   Подсказка: путь относительно текущей директории")
		fmt.Println("   Пример: myFolder, reports/2024")
		inputPath := utils.ReadLine("Directory name: ")
		path := resolveCwd(inputPath)
		err := fs.CreateDirectory(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ Directory created successfully")
			db.LogOperation("create_dir", 0, currentUser.ID)
		}

	case "4":
		fmt.Println("\n📖 Чтение текстового файла")
		fmt.Println("   Подсказка: путь относительно текущей директории")
		fmt.Println("   Пример: test.txt, docs/readme.md")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		content, err := fs.ReadFile(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Content:\n", content)
		}
		db.LogOperation("read_file", 0, currentUser.ID)

	case "5":
		fmt.Println("\n✏️ Запись в текстовый файл")
		fmt.Println("   Подсказка: если файл существует — он будет перезаписан")
		fmt.Println("   Пример: notes.txt, data/info.txt")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		fmt.Println("   Введите содержимое файла (одна строка):")
		content := utils.ReadLine("Content: ")
		err := fs.WriteFile(path, content)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ File written successfully")
			id, _ := db.CreateFileMetadata(inputPath, int64(len(content)), path, currentUser.ID)
			db.LogOperation("write_file", id, currentUser.ID)
		}

	case "6":
		fmt.Println("\n🗑️ Удаление файла")
		fmt.Println("   ⚠️  Внимание: файл будет удалён безвозвратно!")
		fmt.Println("   Пример: old_file.txt, temp/cache.dat")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		err := fs.DeleteFile(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ File deleted")
			db.LogOperation("delete_file", 0, currentUser.ID)
		}

	case "7":
		fmt.Println("\n📋 Копирование файла")
		fmt.Println("   Подсказка: исходный файл останется на месте")
		fmt.Println("   Пример: source.txt → backup/source_copy.txt")
		srcInput := utils.ReadLine("Source path: ")
		dstInput := utils.ReadLine("Dest path: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.CopyFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ File copied")
			db.LogOperation("copy_file", 0, currentUser.ID)
		}

	case "8":
		fmt.Println("\n🚚 Перемещение файла")
		fmt.Println("   Подсказка: файл исчезнет из исходной папки")
		fmt.Println("   Можно использовать для переименования!")
		fmt.Println("   Пример: old.txt → archive/old.txt")
		srcInput := utils.ReadLine("Source path: ")
		dstInput := utils.ReadLine("Dest path: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.MoveFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ File moved")
			db.LogOperation("move_file", 0, currentUser.ID)
		}

	case "9":
		fmt.Println("\n📊 Чтение JSON файла")
		fmt.Println("   Подсказка: файл должен содержать валидный JSON")
		fmt.Println("   Пример: config.json, data/users.json")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		data, err := fs.ReadJSON(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Data: %+v\n", data)
		}
		db.LogOperation("read_json", 0, currentUser.ID)

	case "10":
		fmt.Println("\n📝 Запись JSON файла")
		fmt.Println("   Подсказка: создаёт JSON с одной парой ключ-значение")
		fmt.Println("   Пример: config.json, key=name, value=John")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		key := utils.ReadLine("Key (e.g. username): ")
		val := utils.ReadLine("Value (e.g. admin): ")
		data := map[string]string{key: val}
		err := fs.WriteJSON(path, data)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ JSON written")
			db.LogOperation("write_json", 0, currentUser.ID)
		}

	case "11":
		fmt.Println("\n📄 Чтение XML файла")
		fmt.Println("   Подсказка: файл должен содержать <root><content>...</content></root>")
		fmt.Println("   Пример: data.xml, config/settings.xml")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		data, err := fs.ReadXML(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Data: %+v\n", data)
		}
		db.LogOperation("read_xml", 0, currentUser.ID)

	case "12":
		fmt.Println("\n📝 Запись XML файла")
		fmt.Println("   Подсказка: создаёт XML вида <root><content>ТЕКСТ</content></root>")
		fmt.Println("   Пример: output.xml")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		fmt.Println("   Введите содержимое для тега <content>:")
		content := utils.ReadLine("Content: ")
		data := &fs.XMLData{Content: content}
		err := fs.WriteXML(path, data)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ XML written")
			db.LogOperation("write_xml", 0, currentUser.ID)
		}

	case "13":
		fmt.Println("\n📦 Создание ZIP архива")
		fmt.Println("   Подсказка: можно архивировать файл или целую папку")
		fmt.Println("   Пример: source=docs → dest=docs.zip")
		srcInput := utils.ReadLine("Source Dir/File: ")
		dstInput := utils.ReadLine("Dest Zip path: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.CreateZip(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ Zip created")
			db.LogOperation("create_zip", 0, currentUser.ID)
		}

	case "14":
		fmt.Println("\n📂 Распаковка ZIP архива")
		fmt.Println("   ⚠️  Защита от ZIP-бомб: макс. 100 MB, ratio 100:1")
		fmt.Println("   Пример: archive.zip → extracted/")
		srcInput := utils.ReadLine("Zip path: ")
		dstInput := utils.ReadLine("Dest Dir: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.Unzip(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("✅ Zip extracted")
			db.LogOperation("extract_zip", 0, currentUser.ID)
		}

	case "15":
		currentUser = nil
		fmt.Println("👋 Logged out")

	default:
		fmt.Println("❌ Invalid option. Please enter a number 0-15")
	}
}
