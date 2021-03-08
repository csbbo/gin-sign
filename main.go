package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	jsonFile = "sign.json"
)

type SignData struct {
	Name   string `form:"name"`
	Number string `form:"number"`
}

var globalSingDataList []SignData

func writeJson(singDataList []SignData) error {
	text, err := json.Marshal(singDataList)
	if err != nil {
		return err
	}

	f, err := os.Create(jsonFile)
	if err != nil {
		return err
	}

	defer f.Close()
	f.Write([]byte(string(text)))
	return nil
}

func readJson() ([]SignData, error) {
	var singDataList []SignData
	f, err := os.Open(jsonFile)
	if err != nil {
		return singDataList, nil
	}

	defer f.Close()
	stream, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(stream), &singDataList)
	if err != nil {
		return nil, err
	}
	return singDataList, nil
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
	})

	r.POST("/sign", func(c *gin.Context) {
		var data SignData
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"err": "数据解析失败",
			})
			return
		}

		for _, item := range globalSingDataList {
			if data.Name == item.Name {
				//exists
				c.HTML(http.StatusOK, "sign_success.tmpl", data)
			}
		}

		globalSingDataList = append(globalSingDataList, data)
		err := writeJson(globalSingDataList)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"err": "数据持久化失败",
			})
			return
		}

		c.HTML(http.StatusOK, "sign_success.tmpl", data)
	})

	r.GET("/show", func(c *gin.Context) {
		c.HTML(http.StatusOK, "sign_list.tmpl", globalSingDataList)
	})

	r.GET("/manage", func(c *gin.Context) {
		c.HTML(http.StatusOK, "manage.tmpl", nil)
	})

	r.POST("/empty", func(c *gin.Context) {
		globalSingDataList = nil
		c.HTML(http.StatusOK, "manage.tmpl", gin.H{
			"message": "数据清空成功!",
		})
	})

	r.POST("/import", func(c *gin.Context) {
		var err error
		globalSingDataList, err = readJson()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"err": "数据导入失败",
			})
			return
		}
		c.HTML(http.StatusOK, "manage.tmpl", gin.H{
			"message": "数据从磁盘导入成功!",
		})
	})

	r.POST("/download", func(c *gin.Context) {
		xlsxFile := xlsx.NewFile()
		sheet, err := xlsxFile.AddSheet("Sheet1")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"err": "excel创建失败",
			})
			return
		}

		row := sheet.AddRow()
		cell := row.AddCell()
		cell.Value = "姓名"
		cell = row.AddCell()
		cell.Value = "编号"
		for _, item := range globalSingDataList {
			row := sheet.AddRow()
			cell := row.AddCell()
			cell.Value = item.Name
			cell = row.AddCell()
			cell.Value = item.Number
		}

		stream := new(bytes.Buffer)
		err = xlsxFile.Write(stream)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"err": "excel下载失败",
			})
			return
		}

		c.Header("Access-Control-Expose-Headers", "Content-Disposition")
		c.Header("Content-Disposition", "attachment;filename=sign.xlsx")
		c.Data(http.StatusOK, "application/octet-stream", []byte(stream.String()))
	})

	r.Run("0.0.0.0:8080")
}
