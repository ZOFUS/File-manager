package db

import (
	"log"
)

// LogOperation записывает информацию об операции в журнал аудита
// Все действия пользователей фиксируются в таблице operations
func LogOperation(opType string, fileID int, userID int) {
	stmt, err := DB.Prepare("INSERT INTO operations(operation_type, file_id, user_id) VALUES($1, $2, $3)")
	if err != nil {
		log.Printf("Ошибка подготовки запроса логирования: %v", err)
		return
	}
	defer stmt.Close()

	// Если fileID равен 0, передаём NULL в базу данных
	var fID interface{}
	if fileID == 0 {
		fID = nil
	} else {
		fID = fileID
	}

	_, err = stmt.Exec(opType, fID, userID)
	if err != nil {
		log.Printf("Ошибка логирования операции: %v", err)
	}
}
