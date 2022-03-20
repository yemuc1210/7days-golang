package dialect

import "reflect"

// dialect 关注不同数据库之间的差异部分
// 全局部分
var dialectsMap = map[string]Dialect{}

// Dialect is an interface contains methods that a dialect has to implement
//每个数据库差异的部分称为 dialect
type Dialect interface {
	// 将Go语言类型转换成该数据库的数据类型
	DataTypeOf(typ reflect.Value) string
	// 返回某个表是否存在的SQL语句   不同数据库可能不同
	TableExistSQL(tableName string) (string, []interface{})
}

// RegisterDialect register a dialect to the global variable
// 注册dialect实例
func RegisterDialect(name string, dialect Dialect) {
	// 只要实现了Dialect接口，都可以被接口实例引用
	dialectsMap[name] = dialect
}

// Get the dialect from global variable if it exists
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
