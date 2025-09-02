package handler

import (
	"fmt"
	"reflect"
	"strings"
	"tally-connector/internal/db"
	"tally-connector/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
)

func detDBKeys(d any) []string {
	keys := make([]string, 0)
	v := reflect.ValueOf(d)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.Tag.Get("db") != "" {
			keys = append(keys, field.Tag.Get("db"))
		}
	}
	return keys

}

func FetchGroupHandler(c *gin.Context) {
	groups := []models.MstGroup{}

	keys := detDBKeys(models.MstGroup{})

	stmt := fmt.Sprintf("SELECT %s FROM mst_group ORDER BY sort_position DESC", strings.Join(keys, ", "))

	err := pgxscan.Select(c, db.GetDB(), &groups, stmt)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"result": groups,
		"count":  len(groups),
	})

}
