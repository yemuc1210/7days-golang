package session

import (
	"fmt"
	"geeorm/log"
	"reflect"
	"strings"

	"geeorm/schema"
)

// 存放与数据库表相关的操作

// Model assigns refTable
// 解析操作，由于解析耗时，因此一次性完成解析，将解析结果保存
func (s *Session) Model(value interface{}) *Session {
	// nil or different model, update refTable
	// refTable为nil，说明还未开始解析
	// 或者model发生变化，如结构体名称发生变化，需要重新解析
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		// 将对象实例，解析为数据库表
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// RefTable returns a Schema instance that contains all parsed fields
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		// 若未被赋值，则打印错误日志
		log.Error("Model is not set")
	}
	return s.refTable
}

// 数据库表的操作

// CreateTable create a table in database with a model
func (s *Session) CreateTable() error {
	// 一个模型/对象 对应 一个数据库表
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	// 表 详情描述 ["age int primary key", "name string"]
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

// DropTable drops a table with the name of model
func (s *Session) DropTable() error {
	// Raw构建执行命令
	// Exec 调用sql.DB中的Exec进行执行
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

// HasTable returns true of the table exists
// 判断表是否存在
func (s *Session) HasTable() bool {
	// 由于不同数据库判断表存在的语句不同，所以用到dialect
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	// 根据构建的sql语句，执行查询操作
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().Name
}
