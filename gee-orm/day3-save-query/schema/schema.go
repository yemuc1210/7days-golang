package schema

import (
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

// Dialect实现了一些差异化的SQL语句转换，接下来需要实现核心：对象Object-表table之间的转换
// 给定任何一个对象，将其转换为数据库里的表结构
// 数据库表的要素：表名、字段名和类型、额外约束条件

// Field represents a column of table
// Field 表示一列
// 如：Name string `geeorm:"PRIMARY KEY"`
type Field struct {
	// 字段名
	Name string
	// 字段类型
	Type string
	// 字段标签
	Tag string
}

// Schema represents a table of database
// Schema 表示一个数据库表
type Schema struct {
	// 被映射的对象Model
	Model interface{}
	// 表名
	Name string
	// 字段-对象
	Fields []*Field
	// 所有字段名/列名
	FieldNames []string
	// 字段名和Field之间的映射关系
	// 简化取用操作，不需要遍历Fields
	fieldMap map[string]*Field
}

// GetField returns field by name
func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// Values return the values of dest's member variables
// 新增。从对象中找对应的值，按照数据库表中的顺序平铺
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	// 获取实例
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	// 数据库表中字段的顺序，就是Fields中字段定义的顺序
	for _, field := range schema.Fields {
		// 按照定义的顺序，平铺值
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

// 支持自定义表名
type ITableName interface {
	TableName() string
}

// Parse a struct to a Schema instance
// 将任意的对象解析为Schema实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// 入参dest的值
	//reflect.TypeOf(dest)  这样不行？   效果好像一样‘
	// - 设计的入参是一个对象的指针，因此需要 `reflect.Indirect()` 获取指针指向的实例。
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	var tableName string
	// 是否实现ITableName接口，即是否自定义表名
	t, ok := dest.(ITableName)
	if !ok {
		tableName = modelType.Name()
	} else {
		tableName = t.TableName()
	}
	schema := &Schema{
		Model:    dest,
		Name:     tableName,
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		// 当前要处理的字段
		p := modelType.Field(i)
		// 若p是嵌入的，则是匿名字段，不应被导出
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				// 通过 `(Dialect).DataTypeOf()` 转换为数据库的字段类型
				// Type 是string 类型
				// reflect.Nex()返回的是指针类型，一个特定类型的指针，因此用Indirect获取实例
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			// 处理标签
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			// 更新Scheme的字段和字段名
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
