package db

import (
	"log"
)

func LogOperation(opType string, fileID int, userID int) {
	stmt, err := DB.Prepare("INSERT INTO operations(operation_type, file_id, user_id) VALUES($1, $2, $3)")
	if err != nil {
		log.Printf("Failed to prepare log statement: %v", err)
		return
	}
	defer stmt.Close()

	var fID interface{}
	if fileID == 0 {
		fID = nil
	} else {
		fID = fileID
	}

	_, err = stmt.Exec(opType, fID, userID)
	if err != nil {
		log.Printf("Failed to log operation: %v", err)
	}
}
