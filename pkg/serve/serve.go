package serve

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type TableColumn struct {
	Name     string `yaml:"name" json:"name"`
	Comments string `yaml:"comments" json:"comments"`
	Type     string `yaml:"type" json:"type"`
}

type Table struct {
	Name    string        `yaml:"name" json:"name"`
	Columns []TableColumn `yaml:"columns" json:"columns"`
}

type Dataset struct {
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

type DatasetDocument struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ColumnDocument struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Comments    string `json:"comments"`
	DatasetName string `json:"dataset_name"`
	TableName   string `json:"table_name"`
}

func NewColumnDocument(databaseName, tableName, columnName, columnComments string) *ColumnDocument {
	return &ColumnDocument{
		Id:          fmt.Sprintf("%s.%s.%s", databaseName, tableName, columnName),
		DatasetName: databaseName,
		TableName:   tableName,
		Name:        columnName,
		Comments:    columnComments,
	}
}

type TableDocument struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DatasetName string `json:"dataset_name"`
}

func NewTableDocument(datasetName, tableName string) *TableDocument {
	return &TableDocument{
		Id:          fmt.Sprintf("%s.%s", datasetName, tableName),
		DatasetName: datasetName,
		Name:        tableName,
	}
}

func IndexColumns(datasets []Dataset) (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	// mapping.AddCustomTokenizer(name string, config map[string]interface{})
	mapping.DefaultAnalyzer = simple.Name
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			for _, column := range table.Columns {
				document := NewColumnDocument(
					dataset.Name,
					table.Name,
					column.Name,
					column.Comments,
				)

				if err := index.Index(document.Id, document); err != nil {
					return nil, err
				}
			}
		}
	}

	return index, nil
}

func IndexTables(datasets []Dataset) (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			document := NewTableDocument(
				dataset.Name,
				table.Name,
			)
			if err := index.Index(document.Id, document); err != nil {
				return nil, err
			}
		}
	}

	return index, nil
}

func IndexDataset(datasets []Dataset) (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		for _, table := range dataset.Tables {
			document := NewTableDocument(
				dataset.Name,
				table.Name,
			)
			if err := index.Index(document.Id, document); err != nil {
				return nil, err
			}
		}
	}

	return index, nil
}

type Indices struct {
	DatasetIndex bleve.Index
	TablesIndex  bleve.Index
	ColumnsIndex bleve.Index
}

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

			datasets = append(datasets, Dataset{
				Name:   strings.TrimRight(file.Name(), ".yaml"),
				Tables: tables,
			})
		}
	}

	return datasets, nil
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

func Invoke(configDirectory string, addr string) error {
	datasets, err := BuildDatasets(configDirectory)
	if err != nil {
		return err
	}
	indices, err := BuildIndices(datasets)
	if err != nil {
		return err
	}

	r := gin.Default()
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

		query := bleve.NewQueryStringQuery(q)
		searchRequest := bleve.NewSearchRequestOptions(query, 50, from, false)
		searchRequest.Fields = []string{"*"}
		searchResult, err := indices.ColumnsIndex.Search(searchRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data":    []string{},
				"message": err.Error(),
			})
			return
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

		query := bleve.NewQueryStringQuery(q)
		searchRequest := bleve.NewSearchRequestOptions(query, 50, from, false)
		searchRequest.Fields = []string{"*"}
		searchResult, err := indices.TablesIndex.Search(searchRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"data":    []string{},
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": searchResult,
		})

	})

	return r.Run(addr)
}
