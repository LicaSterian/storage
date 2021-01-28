package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/LicaSterian/storage/api/model"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

// PostUploadFile handler func
func (h Handlers) PostUploadFile(c *gin.Context) {
	res := model.UploadResponse{}
	fileHeader, err := c.FormFile("document")
	if err != nil {
		res.Error = "document not valid"
		c.JSON(http.StatusBadRequest, res)
		log.Println(res.Error, err)
		return
	}

	attachement, err := fileHeader.Open()
	if err != nil {
		res.Error = "cannot open attachement"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}
	defer attachement.Close()

	createdAt := time.Now().UTC().Unix()
	fileID, err := uuid.NewV4()
	if err != nil {
		res.Error = "could not create new uuid"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	tx, err := h.db.BeginTx(context.Background(), nil)
	if err != nil {
		res.Error = "begin tx error"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	_, err = tx.Exec("INSERT INTO files (id, name, created_at, size) VALUES($1, $2, $3, $4)", fileID, fileHeader.Filename, createdAt, fileHeader.Size)
	if err != nil {
		res.Error = "cannot save file to DB"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	file, err := os.Create(fmt.Sprintf("./files/%s", fileID.String()))
	if err != nil {
		res.Error = "cannot create file"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, attachement)
	if err != nil {
		res.Error = "cannot copy attachement to file"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	res.File = model.File{
		ID:        fileID,
		Name:      fileHeader.Filename,
		CreatedAt: createdAt,
		Size:      fileHeader.Size,
	}
	res.Success = true
	c.JSON(http.StatusOK, res)
}

// GetAllFiles handler func
func (h Handlers) GetAllFiles(c *gin.Context) {
	var req model.GetAllRequest

	err := c.ShouldBind(&req)
	if err != nil {
		resErr := model.GetAllResponse{}
		resErr.Error = "cannot bind request body"
		c.JSON(http.StatusBadRequest, resErr)
		log.Println(resErr.Error, err)
		return
	}

	res, statusCode, err := h.storage.GetAll("files", req)
	if err != nil {
		c.JSON(statusCode, res)
		log.Printf("h.storage.GetAll error: %s, req: %+v", res.Error, req)
		return
	}

	c.JSON(statusCode, res)
}

// GetFile handler func
func (h Handlers) GetFile(c *gin.Context) {
	idParam := c.Param("id")
	var f model.File
	var fileID uuid.UUID
	res := model.Response{}
	err := h.db.QueryRow("SELECT name, id, size FROM files WHERE id=$1", idParam).Scan(&f.Name, &fileID, &f.Size)
	switch {
	case err == sql.ErrNoRows:
		res.Error = "file with id not found"
		c.JSON(http.StatusNotFound, res)
		log.Println(res.Error, err)
		return
	case err != nil:
		res.Error = "db queryRow error"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	file, err := os.OpenFile(fmt.Sprintf("./files/%s", fileID.String()), os.O_RDONLY, 0700)
	if err != nil {
		res.Error = "could not open file"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(f.Size, 10))

	_, err = io.Copy(c.Writer, file)
	if err != nil {
		res.Error = "could not copy file"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}
}

// DeleteFile handler func
func (h Handlers) DeleteFile(c *gin.Context) {
	idParam := c.Param("id")
	var fileID uuid.UUID
	res := model.MessageResponse{}
	err := h.db.QueryRow("SELECT id FROM files WHERE id=$1", idParam).Scan(&fileID)
	switch {
	case err == sql.ErrNoRows:
		res.Error = "file with id not found"
		c.JSON(http.StatusNotFound, res)
		log.Println(res.Error, err)
		return
	case err != nil:
		res.Error = "db error"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	tx, err := h.db.BeginTx(context.Background(), nil)
	if err != nil {
		res.Error = "begin tx error"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("DELETE FROM files WHERE id=$1", idParam)
	if err != nil {
		res.Error = "sql exec error"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	err = os.Remove(fmt.Sprintf("./files/%s", fileID.String()))
	if err != nil {
		res.Error = "remove file error"
		c.JSON(http.StatusInternalServerError, res)
		log.Println(res.Error, err)
		return
	}

	res.Success = true
	res.Message = "file deleted"
	c.JSON(http.StatusOK, res)
}
