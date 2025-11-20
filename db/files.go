package db

import (
	"time"
)

type FileMetadata struct {
	ID        int
	Filename  string
	CreatedAt time.Time
	Size      int64
	Location  string
	OwnerID   int
}

func CreateFileMetadata(filename string, size int64, location string, ownerID int) (int, error) {
	stmt, err := DB.Prepare("INSERT INTO files(filename, size, location, owner_id) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(filename, size, location, ownerID).Scan(&id)
	return id, err
}

func GetFileMetadata(id int) (*FileMetadata, error) {
	stmt, err := DB.Prepare("SELECT id, filename, created_at, size, location, owner_id FROM files WHERE id = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var f FileMetadata
	err = stmt.QueryRow(id).Scan(&f.ID, &f.Filename, &f.CreatedAt, &f.Size, &f.Location, &f.OwnerID)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func GetFilesByUser(userID int) ([]FileMetadata, error) {
	stmt, err := DB.Prepare("SELECT id, filename, created_at, size, location, owner_id FROM files WHERE owner_id = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileMetadata
	for rows.Next() {
		var f FileMetadata
		if err := rows.Scan(&f.ID, &f.Filename, &f.CreatedAt, &f.Size, &f.Location, &f.OwnerID); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func DeleteFileMetadata(id int) error {
	stmt, err := DB.Prepare("DELETE FROM files WHERE id = $1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}
