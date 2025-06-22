package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
)

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

//var albums = []album{
//	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
//	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
//	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
//}

var dbConnection *pgx.Conn

func getAlbums(ctx *gin.Context) {
	var albums []album
	rows, err := dbConnection.Query(context.Background(), "SELECT * FROM album")

	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var album album

		if err := rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price); err != nil {
			log.Println(err)
		}
		albums = append(albums, album)
	}

	// Returning albums in JSON with status 200.
	ctx.IndentedJSON(http.StatusOK, albums)
}

func postAlbums(ctx *gin.Context) {
	// Defining newAlbum with type album.
	var newAlbum album

	// Binding JSON to newAlbum, if error - return.
	if err := ctx.BindJSON(&newAlbum); err != nil {
		return
	}

	err := dbConnection.
		QueryRow(context.Background(),
			"INSERT INTO album(title, artist, price) VALUES($1, $2, $3) RETURNING id",
			newAlbum.Title,
			newAlbum.Artist,
			newAlbum.Price,
		).Scan(&newAlbum.ID)

	if err != nil {
		log.Println(err)
	}

	// Appending to in-memory albums newAlbum.
	//albums = append(albums, newAlbum)

	// Returns 201 status and newAlbum in JSON format.
	ctx.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbumByID(ctx *gin.Context) {
	id := ctx.Param("id")

	//var album album
	row := dbConnection.QueryRow(context.Background(), "SELECT * FROM album WHERE id = $1", id)
	var album album
	if err := row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})

			return
		}
		log.Println("Scan error: ", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})

		return
	}

	ctx.IndentedJSON(http.StatusOK, album)
}

func main() {
	dbUrl := "postgres://default:secret@localhost:5432/go"
	var dbErr error
	dbConnection, dbErr = pgx.Connect(context.Background(), dbUrl)
	if dbErr != nil {
		log.Fatal("Connection with DB with error: ", dbErr)
	}
	defer dbConnection.Close(context.Background())

	if ping := dbConnection.Ping(context.Background()); ping != nil {
		log.Fatal("Connection with DB with error: ", dbErr)
	}

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	routerErr := router.Run("localhost:8083")
	if routerErr != nil {
		log.Fatal(routerErr)
	}
}
