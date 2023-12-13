package global

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"os"
	"strings"
	"sync"
)

var (
	GVA_REDIS *redis.Client
	Port      int //端口
	WG        sync.WaitGroup
)

// Column 数据字段类型
type Column struct {
	ColumnName    string `json:"column_name"`
	DataType      string `json:"data_type"`
	ColumnComment string `json:"column_comment"`
	ColumnKey     string `json:"column_key"`
	Extra         string `json:"extra"`
}

// 字符串转为大驼峰
func BigHump(str string) (data string) {
	arr := strings.Split(str, "_")
	for _, a := range arr {
		data += InitialToCapital(a)
	}
	return data
}

// InitialToCapital 首字母转大写
func InitialToCapital(str string) string {
	var InitialToCapitalStr string
	strRune := []rune(str)
	for i := 0; i < len(strRune); i++ {
		if i == 0 {
			if strRune[i] >= 97 && strRune[i] <= 122 {
				strRune[i] -= 32
				InitialToCapitalStr += string(strRune[i])
			} else {
				return str
			}
		} else {
			InitialToCapitalStr += string(strRune[i])
		}
	}
	return InitialToCapitalStr
}

//DBTablesToStructs 数据库-数据表转结构体输出 tables []string{"table_name"}
func DBTablesToStructs(db *gorm.DB, tables []string, model string) {
	if model != "" {
		MkPath("model")
	}
	for _, table := range tables {
		var columns []*Column
		db.Debug().Raw("select column_name, data_type, column_comment, column_key, extra from information_schema.columns where table_name = ? and table_schema =(select database()) order by ordinal_position ", table).Scan(&columns)
		TableToStruct(columns, table, model)
	}
}

// 数据表转结构体
func TableToStruct(data []*Column, table, model string) {
	// ----- 拼接生成的struct  start--------
	bh_table := BigHump(table)
	structStr := ""
	if model != "" {
		structStr = fmt.Sprintf("package %s\n\ntype %s struct {\n", model, bh_table)
	} else {
		structStr = fmt.Sprintf("type %s struct {\n", bh_table)
	}
	for _, column := range data {
		structStr += "    " + BigHump(column.ColumnName) //InitialToCapital(column.ColumnName)
		if column.DataType == "tinyint" {
			structStr += " int "
		} else if column.DataType == "decimal" {
			structStr += " float64 "
		} else if column.DataType == "bigint" || column.DataType == "int" {
			structStr += " int64 "
		} else {
			structStr += " string "
		}
		structStr += fmt.Sprintf("`gorm:\"column:%s;comment('%s')\" json:\"%s\"` \n", column.ColumnName, column.ColumnComment, column.ColumnName)
		//if column.Extra != "auto_increment" {
		//	structStr += fmt.Sprintf("`gorm:\"comment('%s')\" json:\"%s\"` \n",
		//		column.ColumnComment, column.ColumnName)
		//} else {
		//	structStr += fmt.Sprintf("`gorm:\"not null comment('%s') INT(11)\" json:\"%s\"` \n",
		//		column.ColumnComment, column.ColumnName)
		//}
	}
	structStr += "}\n\n"
	// 拼接 方法 TableName 返回table名称
	structStr += "func (" + bh_table + ") TableName() string {\n\treturn \"" + table + "\"\n}"
	if model != "" {
		if !MkFile(model+"/"+table+".go", structStr) {
			fmt.Println("写入失败")
			fmt.Println(structStr)
		}
	} else {
		fmt.Println(structStr)
	}
}

func Sha256(str string) string {
	m := sha256.New()
	m.Write([]byte(str))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

func MkPath(path string) bool {
	if !PathExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("path err:", err)
			return false
		}
		return true
	}
	return true
}

func MkFile(path, info string) bool {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("file err:", err)
		return false
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(info)
	if err != nil {
		fmt.Println("file write err:", err)
		return false
	}
	writer.Flush()
	return true
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func StrsToInterfaces(strs []string, l int) []interface{} {
	if l == 0 {
		l = len(strs)
	}
	ins := make([]interface{}, l)
	for i, str := range strs {
		ins[i] = str
		if i == l {
			break
		}
	}
	return ins
}
