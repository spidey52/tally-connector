package main

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"tally-connector/cmd/db"
	"tally-connector/cmd/helper"
	"tally-connector/cmd/loader/config"
	"tally-connector/cmd/models"
	"time"

	"github.com/gin-gonic/gin"
)

// var table = "trn_employee"

func getTableConfig(table string) *config.Table {
	var tablesConfigs *config.TablesConfig

	var err error
	if tablesConfigs, err = config.LoadTablesConfig(); err != nil {
		return nil
	}

	for _, tbl := range tablesConfigs.Master {
		if tbl.Name == table {
			return &tbl
		}
	}

	for _, tbl := range tablesConfigs.Transaction {
		if tbl.Name == table {
			return &tbl
		}
	}

	return nil
}

func parseDate(str string) time.Time {
	layout := "2006-01-02"
	t, err := time.Parse(layout, str)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		return time.Time{}
	}
	return t
}

var client = &http.Client{}

// var tallyEndpoint = "http://100.77.107.9:9000"
// var tallyEndpoint = "http://172.16.0.47:9000"
var tallyEndpoint = "http://100.65.128.68:9000"

func callTallyApi(ctx context.Context, xmlData []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tallyEndpoint, bytes.NewReader(xmlData))
	if err != nil {
		return "", err
	}
	curr_time := time.Now().Format("20060102150405")
	os.WriteFile(fmt.Sprintf("request_%s.xml", curr_time), xmlData, 0644)

	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	if err != nil {

		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Tally API response status: %s", resp.Status)
		return "", fmt.Errorf("failed to call Tally API: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Error cleaning response: %v", err)
		return "", err
	}

	os.WriteFile(fmt.Sprintf("response_%s.xml", curr_time), []byte(string(body)), 0644)
	str := helper.CleanString(string(body))

	return str, nil
}

func importData(ctx context.Context, import_table models.SyncTable) (string, error) {
	var cfg *config.Config

	var err error
	if cfg, err = config.LoadConfig(); err != nil {
		return "", err
	}

	configTallyXML := make(map[string]any)
	configTallyXML["fromDate"] = parseDate(cfg.Tally.FromDate)
	configTallyXML["toDate"] = parseDate(cfg.Tally.ToDate)
	if cfg.Tally.Company != "" {
		configTallyXML["targetCompany"] = html.EscapeString(cfg.Tally.Company)
	} else {
		configTallyXML["targetCompany"] = "##SVCurrentCompany"
	}

	table := getTableConfig(import_table.Name)

	if table == nil {
		return "", fmt.Errorf("table config not found for %s", import_table.Name)
	}

	tblConfig := TableConfigYAML{
		Name:       table.Name,
		Collection: table.Collection,
		Fields:     table.Fields,
		Fetch:      table.Fetch,
		Filters:    table.Filters,
	}

	// Generate XML from YAML
	result, _ := GenerateXMLfromYAML(tblConfig, TallyConfig{
		Company:  cfg.Tally.Company,
		FromDate: cfg.Tally.FromDate,
		ToDate:   cfg.Tally.ToDate,
	})

	start := time.Now()

	// TODO: use the max_wait from the table config
	reqCtx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	val, err := callTallyApi(reqCtx, []byte(result))

	if err != nil {
		log.Printf("Error calling Tally API: %v", err.Error())
		log.Println("Failed to import data for", table.Name, err.Error())
		return "", err
	}

	log.Println("Tally API call duration:", time.Since(start).Seconds(), "seconds", "for", table.Name)
	fmt.Println("Response length:", len(val), "for", table.Name)
	return val, nil
}

func main() {
	// db.ConnectDB("postgres://satyam:satyam52@localhost:5432/tally_incremental_db")
	db.ConnectDB("postgres://satyam:satyam52@100.66.94.61:5432/tally_db")
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()

	server.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		table_name := c.Query("table_name")

		// q := psql.Select(
		// 	sm.Columns("table_name", "group_name", "max_wait", "sync_interval"),
		// 	sm.From("tbl_sync_tables"),
		// 	sm.Where(psql.Quote("table_name").EQ(psql.Arg(table_name))),
		// 	sm.Limit(1),
		// )

		// query, args, err := q.Build(ctx)

		// if err != nil {
		// 	c.JSON(500, gin.H{"error": "Failed to build query", "details": err.Error()})
		// 	return
		// }

		var tables = []models.SyncTable{}

		// err = pgxscan.Select(ctx, db.GetDB(), &tables, query, args...)

		// if err != nil {
		// 	c.JSON(500, gin.H{"error": "Failed to execute query", "details": err.Error()})
		// 	return
		// }

		// if len(tables) == 0 {
		// 	c.JSON(400, gin.H{"error": "No tables found"})
		// 	return
		// }

		loadTables, err := config.LoadTablesConfig()

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to load table config", "details": err.Error()})
			return
		}

		for _, table := range loadTables.Master {
			if table.Name == table_name {
				tables = append(tables, models.SyncTable{
					Name: table.Name,
				})
			}

		}

		for _, table := range loadTables.Transaction {
			if table.Name == table_name {
				tables = append(tables, models.SyncTable{
					Name: table.Name,
				})

				fmt.Println("Found transaction table:", table.Fetch)

			}
		}

		result, err := importData(ctx, tables[0])

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to import data", "details": err.Error()})
			return
		}

		if len(result) > 0 {
			tableName := tables[0].Name

			tableConfig := getTableConfig(tableName)

			if tableConfig == nil {
				c.JSON(404, gin.H{"error": "Table configuration not found"})
				return
			}

			ProcessXMLData(c, tableName, result, tableConfig.Fields)

			return
		}

		c.JSON(200, gin.H{
			"data":  result,
			"count": len(result),
		})
	})

	log.Println("Starting server on :8081")
	server.Run(":8081")
}
