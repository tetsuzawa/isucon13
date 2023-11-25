package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// mysql
	_ "github.com/go-sql-driver/mysql"
	// postgres
	//_ "github.com/mackee/pgx-replaced"
)

type Icon struct {
	ID       int64  `db:"id"`
	UserID   int64  `db:"user_id"`
	UserName string `db:"user_name"`
	Image    []byte `db:"image"`
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Ltime)

	exportDir := flag.String("export-dir", "/home/isucon/webapp/public/img/", "export directory")
	flag.Parse()

	db, err := GetDBPostgres()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve image data from the database
	var icons []Icon
	query := `SELECT icons.id as id, name as user_name , image FROM icons JOIN isupipe.users on icons.user_id = users.id`
	err = db.Select(&icons, query)
	if err != nil {
		log.Fatal(err)
	}

	// Save each image to a file
	for _, icon := range icons {
		filename := filepath.Join(*exportDir, fmt.Sprintf("%d.jpeg", icon.UserName))
		err := os.WriteFile(filename, icon.Image, 0777)
		if err != nil {
			log.Printf("Failed to write file for icon ID %d: %v\n", icon.ID, err)
			continue
		}
		fmt.Printf("File saved: %s\n", filename)
	}
}

// mimeToExt converts a MIME type to a file extension
