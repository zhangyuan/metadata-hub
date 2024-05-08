package serve

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/token/ngram"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/letter"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"metadata-hub/ui"
)

type TableColumn struct {
	Id       string
	Name     string `yaml:"name" json:"name"`
	Comments string `yaml:"comments" json:"comments"`
	Type     string `yaml:"type" json:"type"`
}

type Table struct {
	Id       string
	Name     string        `yaml:"name" json:"name"`
	Comments string        `yaml:"comments" json:"comments"`
	Columns  []TableColumn `yaml:"columns" json:"columns"`
}

type Dataset struct {
	Id     string
	Name   string
	Tables []Table
}

func Parse[T interface{}](source []byte) ([]T, error) {
	dec := yaml.NewDecoder(bytes.NewReader(source))

	documents := []T{}
	for {
		var document T
		if err := dec.Decode(&document); err != nil {
			if err == io.EOF {
				return documents, nil
			}
			return nil, err
		}

		documents = append(documents, document)
	}
}

type ColumnDocument struct {
	Document string `json:"document"`
}

func NewColumnDocument(dataset *Dataset, table *Table, column *TableColumn) *ColumnDocument {
	return &ColumnDocument{
		Document: strings.Join([]string{column.Name, column.Comments}, "\n"),
	}
}

type TableDocument struct {
	Document string `json:"document"`
}

func NewTableDocument(dataset *Dataset, table *Table) *TableDocument {
	return &TableDocument{
		Document: strings.Join([]string{table.Name, table.Comments}, "\n"),
	}
}

func IndexColumns(datasets []Dataset) (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	if err := mapping.AddCustomTokenFilter("content_ngram", map[string]interface{}{
		"type": ngram.Name,
		"min":  1,
		"max":  3,
	}); err != nil {
		return nil, err
	}

	if err := mapping.AddCustomAnalyzer("custom", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			lowercase.Name,
			"content_ngram",
		},
	}); err != nil {
		return nil, err
	}

	mapping.DefaultAnalyzer = "custom"

	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			for _, column := range table.Columns {
				document := NewColumnDocument(&dataset, &table, &column)
				if err := index.Index(column.Id, document); err != nil {
					return nil, err
				}
			}
		}
	}

	return index, nil
}

func IndexTables(datasets []Dataset) (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()

	if err := mapping.AddCustomTokenFilter("content_ngram", map[string]interface{}{
		"type": ngram.Name,
		"min":  1,
		"max":  3,
	}); err != nil {
		return nil, err
	}

	if err := mapping.AddCustomAnalyzer("custom", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": letter.Name,
		"token_filters": []string{
			lowercase.Name,
			"content_ngram",
		},
	}); err != nil {
		return nil, err
	}

	mapping.DefaultAnalyzer = "custom"

	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			document := NewTableDocument(&dataset, &table)
			if err := index.Index(table.Id, document); err != nil {
				return nil, err
			}
		}
	}

	return index, nil
}

type Indices struct {
	TablesIndex  bleve.Index
	ColumnsIndex bleve.Index
}

type TableColumnRecord struct {
	Id          string
	Name        string
	Comments    string
	DatasetName string
	TableName   string
}

type TableRecord struct {
	Id          string
	Name        string
	Comments    string
	DatasetName string
}

type ColumnStore map[string]TableColumnRecord
type TableStore map[string]TableRecord

type Store struct {
	ColumnStore ColumnStore
	TableStore  TableStore
}

const sep = "/"

func BuildDatasets(configDirectory string) ([]Dataset, error) {
	files, err := os.ReadDir(configDirectory)
	if err != nil {
		return nil, err
	}

	datasets := []Dataset{}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			path := filepath.Join(configDirectory, file.Name())
			bytes, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			tables, err := Parse[Table](bytes)

			if err != nil {
				return nil, err
			}

			datasetName := strings.TrimRight(file.Name(), ".yaml")
			for tableIdx := range tables {
				table := &tables[tableIdx]
				table.Id = strings.Join([]string{datasetName, table.Name}, sep)

				for columnIdx := range table.Columns {
					column := &table.Columns[columnIdx]
					column.Id = strings.Join([]string{datasetName, table.Id, column.Name}, sep)
				}
			}

			datasets = append(datasets, Dataset{
				Id:     datasetName,
				Name:   datasetName,
				Tables: tables,
			})
		}
	}

	return datasets, nil
}

func BuildStore(datasets []Dataset) *Store {
	columnStore := ColumnStore{}
	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			for _, column := range table.Columns {
				columnStore[column.Id] = TableColumnRecord{
					Id:          column.Id,
					Name:        column.Name,
					Comments:    column.Comments,
					DatasetName: dataset.Name,
					TableName:   table.Name,
				}
			}
		}
	}

	tableStore := TableStore{}
	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			tableStore[table.Id] = TableRecord{
				Id:          table.Id,
				Name:        table.Name,
				Comments:    table.Comments,
				DatasetName: dataset.Name,
			}
		}
	}

	return &Store{
		TableStore:  tableStore,
		ColumnStore: columnStore,
	}
}

func BuildIndices(datasets []Dataset) (*Indices, error) {
	columnsIndex, err := IndexColumns(datasets)
	if err != nil {
		return nil, err
	}
	tablesIndex, err := IndexTables(datasets)
	if err != nil {
		return nil, err
	}

	return &Indices{
		ColumnsIndex: columnsIndex,
		TablesIndex:  tablesIndex,
	}, nil
}

func BuildUIAssetPath(path string) string {
	return fmt.Sprintf("dist%s", path)
}

func Invoke(configDirectory string, addr string) error {
	datasets, err := BuildDatasets(configDirectory)
	if err != nil {
		return err
	}
	indices, err := BuildIndices(datasets)
	if err != nil {
		return err
	}

	store := BuildStore(datasets)

	if strings.ToLower(os.Getenv("GIN_MODE")) == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	staticFS := ui.StaticFS

	r.Use(func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.Path

		if strings.HasPrefix(requestPath, "/api/") {
			ctx.Next()
			return
		}
		if ctx.Request.Method == "GET" {
			extentions := []string{".css", ".js", ".ico", ".png", ".jpg", ".svg"}
			for _, extension := range extentions {
				if strings.HasSuffix(requestPath, extension) {
					data, err := staticFS.ReadFile(BuildUIAssetPath(requestPath))
					if err != nil {
						_ = ctx.AbortWithError(500, err)
						return
					} else {
						ctx.Data(200, mime.TypeByExtension(extension), data)
						return
					}
				}
			}

			acceptHeader := ctx.Request.Header.Get("Accept")
			if strings.Contains(acceptHeader, "text/html") || strings.Contains(acceptHeader, "*/*") {
				file, err := staticFS.ReadFile(BuildUIAssetPath("/index.html"))
				if err != nil {
					_ = ctx.AbortWithError(500, err)
					return
				}
				ctx.Data(200, "text/html", file)
				return
			}

			return
		}
		ctx.AbortWithStatus(404)
	})

	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/api/datasets", func(c *gin.Context) {
		dataseNames := []string{}
		for _, dataset := range datasets {
			dataseNames = append(dataseNames, dataset.Name)
		}

		c.JSON(http.StatusOK, gin.H{
			"data": dataseNames,
		})
	})

	r.GET("/api/datasets/:datasetName", func(c *gin.Context) {
		datasetName := c.Params.ByName("datasetName")
		tableNames := []string{}
		for _, dataset := range datasets {
			if dataset.Name == datasetName {
				for _, table := range dataset.Tables {
					tableNames = append(tableNames, table.Name)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": map[string]interface{}{
				"tables": tableNames,
			},
		})
	})

	r.GET("/api/datasets/:datasetName/tables/:tableName", func(c *gin.Context) {
		datasetName := c.Params.ByName("datasetName")
		tableName := c.Params.ByName("tableName")

		var targetTable Table
		for _, dataset := range datasets {
			if dataset.Name == datasetName {
				for _, table := range dataset.Tables {
					if table.Name == tableName {
						targetTable = table
					}
				}
			}

		}

		c.JSON(http.StatusOK, gin.H{
			"data": targetTable,
		})
	})

	r.GET("/api/search-columns", func(c *gin.Context) {
		q := c.Query("q")
		fromStr := c.Query("from")
		from, err := strconv.Atoi(fromStr)
		if err != nil {
			from = 0
		}

		if q == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": []string{},
			})
			return
		}

		documentMatchQuery := bleve.NewMatchQuery(q)
		documentMatchQuery.SetOperator(query.MatchQueryOperatorAnd)
		documentMatchQuery.SetField("document")

		searchRequest := bleve.NewSearchRequestOptions(documentMatchQuery, 100, from, false)

		searchRequest.Fields = []string{"*"}
		searchResult, err := indices.ColumnsIndex.Search(searchRequest)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data":    []string{},
				"message": err.Error(),
			})
			return
		}

		for hitIdx := range searchResult.Hits {
			hit := searchResult.Hits[hitIdx]
			columnRecord := store.ColumnStore[hit.ID]
			searchResult.Hits[hitIdx].Fields["datasetName"] = columnRecord.DatasetName
			searchResult.Hits[hitIdx].Fields["tableName"] = columnRecord.TableName
			searchResult.Hits[hitIdx].Fields["name"] = columnRecord.Name
			searchResult.Hits[hitIdx].Fields["comments"] = columnRecord.Comments

			fmt.Println(hit.ID, columnRecord.Id)
		}

		c.JSON(http.StatusOK, gin.H{
			"data": searchResult,
		})
	})

	r.GET("/api/search-tables", func(c *gin.Context) {
		q := c.Query("q")
		fromStr := c.Query("from")
		from, err := strconv.Atoi(fromStr)
		if err != nil {
			from = 0
		}

		if q == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": []string{},
			})
			return
		}

		documentMatchQuery := bleve.NewMatchQuery(q)
		documentMatchQuery.SetOperator(query.MatchQueryOperatorAnd)
		documentMatchQuery.SetField("document")

		searchRequest := bleve.NewSearchRequestOptions(documentMatchQuery, 100, from, false)

		searchRequest.Fields = []string{"*"}
		searchResult, err := indices.TablesIndex.Search(searchRequest)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data":    []string{},
				"message": err.Error(),
			})
			return
		}

		for hitIdx := range searchResult.Hits {
			hit := searchResult.Hits[hitIdx]
			tableRecord := store.TableStore[hit.ID]
			hit.Fields["datasetName"] = tableRecord.DatasetName
			hit.Fields["name"] = tableRecord.Name
			hit.Fields["comments"] = tableRecord.Comments
		}

		c.JSON(http.StatusOK, gin.H{
			"data": searchResult,
		})

	})

	return r.Run(addr)
}
