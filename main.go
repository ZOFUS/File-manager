package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	fmt.Println("\n────────────── Main Menu ──────────────")
	fmt.Printf("User: %s | Dir: /%s\n", currentUser.Username, currentDir)
	fmt.Println("────────────────────────────────────────")
	fmt.Println("НАВИГАЦИЯ")
	fmt.Println("   1. Перейти в папку (cd)")
	fmt.Println("   2. Показать содержимое папки")
	fmt.Println("   3. Создать папку")
	fmt.Println("   4. Информация о дисках")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("ФАЙЛЫ")
	fmt.Println("   5. Создать/записать файл")
	fmt.Println("   6. Прочитать файл")
	fmt.Println("   7. Редактировать файл")
	fmt.Println("   8. Удалить файл")
	fmt.Println("   9. Копировать файл")
	fmt.Println("  10. Переместить файл")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("ДАННЫЕ (JSON/XML)")
	fmt.Println("  11. Создать JSON    12. Прочитать JSON")
	fmt.Println("  13. Создать XML     14. Прочитать XML")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("АРХИВЫ")
	fmt.Println("  15. Создать ZIP     16. Распаковать ZIP")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("   0. Выход (Logout)")

	choice := utils.ReadLine("Select option: ")

	switch choice {
	// ==================== НАВИГАЦИЯ ====================
	case "1": // Перейти в папку (cd)
		fmt.Println("\nСмена текущей директории")
		fmt.Println("   Подсказка: введите '..' для перехода наверх")
		fmt.Println("   Пример: docs, .., subdir, / (корень sandbox)")
		newDir := utils.ReadLine("New directory: ")
		if err := changeDirectory(newDir); err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("OK. Перешли в: /%s\n", currentDir)
		}

	case "2": // Показать содержимое папки
		fmt.Println("\nПросмотр содержимого директории")
		fmt.Println("   Пример: . (текущая), docs, subdir/nested")
		inputPath := utils.ReadLine("Path [. = current]: ")
		path := resolveCwd(inputPath)
		files, err := fs.ListDirectory(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("\nСодержимое /%s:\n", path)
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

	case "3": // Создать папку
		fmt.Println("\nСоздание новой директории")
		fmt.Println("   Пример: myFolder, reports/2024")
		inputPath := utils.ReadLine("Directory name: ")
		path := resolveCwd(inputPath)
		err := fs.CreateDirectory(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. Directory created")
			db.LogOperation("create_dir", 0, currentUser.ID)
		}

	case "4": // Информация о дисках
		fmt.Println("\nИнформация о дисках/файловой системе")
		drives := fs.ListDrives()
		fmt.Println("Доступные разделы:", drives)
		diskInfo, err := fs.GetDiskInfo("/")
		if err == nil {
			fmt.Printf("\nРаздел: %s\n", diskInfo.Name)
			fmt.Printf("   Всего:     %.2f GB\n", float64(diskInfo.TotalSize)/(1024*1024*1024))
			fmt.Printf("   Свободно:  %.2f GB\n", float64(diskInfo.FreeSpace)/(1024*1024*1024))
			fmt.Printf("   Занято:    %.2f GB (%.1f%%)\n", float64(diskInfo.UsedSpace)/(1024*1024*1024), diskInfo.UsedPercent)
		} else {
			fmt.Println("   Не удалось получить информацию о диске:", err)
		}
		db.LogOperation("list_drives", 0, currentUser.ID)

	// ==================== ФАЙЛЫ ====================
	case "5": // Создать/записать файл
		fmt.Println("\nЗапись в текстовый файл")
		fmt.Println("   Если файл существует — он будет перезаписан")
		fmt.Println("   Пример: notes.txt, data/info.txt")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		fmt.Println("   Введите содержимое файла:")
		content := utils.ReadLine("Content: ")
		err := fs.WriteFile(path, content)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. File written")
			id, _ := db.CreateFileMetadata(inputPath, int64(len(content)), path, currentUser.ID)
			db.LogOperation("write_file", id, currentUser.ID)
		}

	case "6": // Прочитать файл
		fmt.Println("\nЧтение текстового файла")
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

	case "7": // Редактировать файл
		fmt.Println("\nРедактирование файла")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		currentContent, err := fs.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		// Разбиваем на строки
		lines := strings.Split(currentContent, "\n")

		fmt.Println("\n────────────────────────────────")
		fmt.Println("Содержимое файла (по строкам):")
		fmt.Println("────────────────────────────────")
		for i, line := range lines {
			fmt.Printf("  %d: %s\n", i+1, line)
		}
		fmt.Println("────────────────────────────────")

		fmt.Println("\nВыберите действие:")
		fmt.Println("1. Редактировать строку")
		fmt.Println("2. Добавить строку в конец")
		fmt.Println("3. Удалить строку")
		fmt.Println("4. Перезаписать всё")
		fmt.Println("0. Отмена")
		action := utils.ReadLine("Действие: ")

		switch action {
		case "1": // Редактировать строку
			lineNumStr := utils.ReadLine("Номер строки для редактирования: ")
			lineNum := 0
			fmt.Sscanf(lineNumStr, "%d", &lineNum)
			if lineNum < 1 || lineNum > len(lines) {
				fmt.Println("Неверный номер строки")
				return
			}
			fmt.Printf("Текущее значение: %s\n", lines[lineNum-1])
			newLine := utils.ReadLine("Новое значение: ")
			lines[lineNum-1] = newLine
			newContent := strings.Join(lines, "\n")
			err = fs.WriteFile(path, newContent)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("OK. Строка изменена")
				db.LogOperation("edit_file", 0, currentUser.ID)
			}
		case "2": // Добавить строку
			newLine := utils.ReadLine("Новая строка: ")
			err = fs.AppendFile(path, "\n"+newLine)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("OK. Строка добавлена")
				db.LogOperation("edit_file", 0, currentUser.ID)
			}
		case "3": // Удалить строку
			lineNumStr := utils.ReadLine("Номер строки для удаления: ")
			lineNum := 0
			fmt.Sscanf(lineNumStr, "%d", &lineNum)
			if lineNum < 1 || lineNum > len(lines) {
				fmt.Println("Неверный номер строки")
				return
			}
			lines = append(lines[:lineNum-1], lines[lineNum:]...)
			newContent := strings.Join(lines, "\n")
			err = fs.WriteFile(path, newContent)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("OK. Строка удалена")
				db.LogOperation("edit_file", 0, currentUser.ID)
			}
		case "4": // Перезаписать всё
			fmt.Println("Введите новое содержимое:")
			newContent := utils.ReadLine("Content: ")
			err = fs.WriteFile(path, newContent)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("OK. Файл перезаписан")
				db.LogOperation("edit_file", 0, currentUser.ID)
			}
		case "0":
			fmt.Println("Отменено")
		default:
			fmt.Println("Неверное действие")
		}

	case "8": // Удалить файл
		fmt.Println("\nУдаление файла")
		fmt.Println("   Внимание: файл будет удалён безвозвратно!")
		fmt.Println("   Пример: old_file.txt, temp/cache.dat")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		err := fs.DeleteFile(path)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. File deleted")
			db.LogOperation("delete_file", 0, currentUser.ID)
		}

	case "9": // Копировать файл
		fmt.Println("\nКопирование файла")
		fmt.Println("   Исходный файл останется на месте")
		fmt.Println("   Пример: source.txt -> backup/source_copy.txt")
		srcInput := utils.ReadLine("Source path: ")
		dstInput := utils.ReadLine("Dest path: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.CopyFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. File copied")
			db.LogOperation("copy_file", 0, currentUser.ID)
		}

	case "10": // Переместить файл
		fmt.Println("\nПеремещение файла")
		fmt.Println("   Файл исчезнет из исходной папки")
		fmt.Println("   Можно использовать для переименования!")
		fmt.Println("   Пример: old.txt -> archive/old.txt")
		srcInput := utils.ReadLine("Source path: ")
		dstInput := utils.ReadLine("Dest path: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.MoveFile(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. File moved")
			db.LogOperation("move_file", 0, currentUser.ID)
		}

	// ==================== ДАННЫЕ (JSON/XML) ====================
	case "11": // Создать JSON
		fmt.Println("\nЗапись JSON файла")
		fmt.Println("   Введите любой валидный JSON")
		fmt.Println("   Пример: {\"name\": \"John\", \"age\": 25}")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		fmt.Println("   Введите JSON:")
		jsonContent := utils.ReadLine("JSON: ")
		err := fs.WriteFile(path, jsonContent)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. JSON файл создан")
			db.LogOperation("write_json", 0, currentUser.ID)
		}

	case "12": // Прочитать JSON
		fmt.Println("\nЧтение JSON файла")
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

	case "13": // Создать XML
		fmt.Println("\nЗапись XML файла")
		fmt.Println("   Введите любой валидный XML")
		fmt.Println("   Пример: <user><name>John</name></user>")
		inputPath := utils.ReadLine("File path: ")
		path := resolveCwd(inputPath)
		fmt.Println("   Введите XML:")
		xmlContent := utils.ReadLine("XML: ")
		err := fs.WriteFile(path, xmlContent)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. XML файл создан")
			db.LogOperation("write_xml", 0, currentUser.ID)
		}

	case "14": // Прочитать XML
		fmt.Println("\nЧтение XML файла")
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

	// ==================== АРХИВЫ ====================
	case "15": // Создать ZIP
		fmt.Println("\nСоздание ZIP архива")
		fmt.Println("   Шаг 1: укажите ЧТО архивировать (файл или папку)")
		fmt.Println("   Шаг 2: укажите ИМЯ архива (например: archive.zip)")
		srcInput := utils.ReadLine("Что архивировать: ")
		dstInput := utils.ReadLine("Имя архива (.zip): ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.CreateZip(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. Zip created")
			db.LogOperation("create_zip", 0, currentUser.ID)
		}

	case "16": // Распаковать ZIP
		fmt.Println("\nРаспаковка ZIP архива")
		fmt.Println("   Шаг 1: укажите ZIP файл")
		fmt.Println("   Шаг 2: укажите ПАПКУ для распаковки")
		srcInput := utils.ReadLine("ZIP файл: ")
		dstInput := utils.ReadLine("Папка назначения: ")
		src := resolveCwd(srcInput)
		dst := resolveCwd(dstInput)
		err := fs.Unzip(src, dst)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("OK. Zip extracted")
			db.LogOperation("extract_zip", 0, currentUser.ID)
		}

	// ==================== ВЫХОД ====================
	case "0":
		currentUser = nil
		fmt.Println("Logged out")

	default:
		fmt.Println("Invalid option")
	}
}
