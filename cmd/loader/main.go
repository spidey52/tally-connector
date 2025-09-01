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
	"slices"
	"tally-connector/cmd/loader/config"
	"tally-connector/internal/db"
	"tally-connector/internal/helper"
	"tally-connector/internal/models"
	"time"

	"github.com/joho/godotenv"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
)

type LoaderEnv struct {
	PostgresURL   string `env:"POSTGRES_URL"`
	TallyEndpoint string `env:"TALLY_ENDPOINT"`
}

var env LoaderEnv

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	env.PostgresURL = os.Getenv("POSTGRES_URL")
	env.TallyEndpoint = os.Getenv("TALLY_ENDPOINT")
}

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

func callTallyApi(ctx context.Context, xmlData []byte) (string, error) {
	var tallyEndpoint = env.TallyEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tallyEndpoint, bytes.NewReader(xmlData))
	if err != nil {
		return "", err
	}
	curr_time := time.Now().Format("20060102150405")
	os.WriteFile(fmt.Sprintf("reqlog/request/request_%s.xml", curr_time), xmlData, 0644)

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

	str := helper.CleanString(string(body))
	os.WriteFile(fmt.Sprintf("reqlog/response/response_%s.xml", curr_time), []byte(string(body)), 0644)

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

	// TODO: use the max_wait from the table config
	reqCtx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	val, err := callTallyApi(reqCtx, []byte(result))

	if err != nil {
		log.Printf("Error calling Tally API: %v", err.Error())
		log.Println("Failed to import data for", table.Name, err.Error())
		return "", err
	}

	err = ProcessXMLData(ctx, table.Name, val, table.Fields)

	if err != nil {
		return "", err
	}

	return val, nil
}

func ImportAll(filters ...string) {
	tables, err := config.GetMergedTables()

	if err != nil {
		log.Printf("Error loading table config: %v", err)
		return
	}

	for _, table := range tables {
		if len(filters) > 0 && !slices.Contains(filters, table.Name) {
			continue
		}

		start := time.Now()
		_, err = importData(context.Background(), models.SyncTable{
			Name: table.Name,
		})
		// duration := time.Since(start)

		var res = models.SyncLog{
			Table:     table.Name,
			Group:     table.Name,
			StartTime: start,
			EndTime:   time.Now(),
			Duration:  time.Since(start).Seconds(),
			Status:    "success",
			Message:   "",
		}

		if err != nil {
			res.Status = "failed"
			res.Message = err.Error()
		}

		q := psql.Insert(
			im.Into("tbl_sync_logs"),
			im.Values(
				psql.Arg(res.Table, res.Group, res.StartTime, res.EndTime, res.Duration, res.Status, res.Message),
			),
		)

		query, args, err := q.Build(context.Background())

		if err != nil {
			log.Printf("Error building query: %v", err)
			return
		}

		_, err = db.GetDB().Exec(context.Background(), query, args...)
		if err != nil {
			log.Printf("Error executing query: %v", err)
			return
		}

	}

}

func main() {
	db.ConnectDB(env.PostgresURL)
	ProcessImportQueue()

	// gin.SetMode(gin.ReleaseMode)
	// server := gin.Default()

	// server.POST("/seed-tables", func(c *gin.Context) {
	// 	models.CreateSyncTables()
	// 	SeedSyncTables(context.Background())
	// 	c.JSON(200, gin.H{"status": "tables created"})
	// })

	// server.POST("/sync", func(c *gin.Context) {
	// 	type SyncDto struct {
	// 		Filters []string `json:"filters"`
	// 	}

	// 	var dto SyncDto
	// 	if err := c.ShouldBindJSON(&dto); err != nil {
	// 		c.JSON(400, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	if len(dto.Filters) == 0 {
	// 		c.JSON(400, gin.H{"error": "no filters provided", "message": "please provide at least one filter"})
	// 		return
	// 	}

	// 	redisclient.GetRedisClient().LPush(context.Background(), redisclient.ImportQueueKey, dto.Filters)

	// 	c.JSON(200, gin.H{
	// 		"status": "added in import queue",
	// 	})
	// })

	// log.Println("Starting server on :8081")
	// server.Run(":8081")
}
