package clause

import (
	"strings"
)

// Clause: 子句
// 查询语句可以包含许多子句
// 构造完整SQL语句放在Clause包

// Clause contains SQL conditions
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

// Type is the type of Clause
type Type int

// Support types for Clause
// 定义Clause支持的类型
const (
	INSERT  Type = iota // 0
	VALUES              // 1
	SELECT              // 2
	LIMIT               // 3
	WHERE               // 4
	ORDERBY             // 5
)

// Set adds a sub clause of specific type
// Set方法根据Type 调用对应的generator，生成该子句对应的SQL语句
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	// 产生子句
	sql, vars := generators[name](vars...)
	// 存入结构体
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// Build generate the final SQL and SQLVars
// 根据子句，构建最终的SQL语句
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	// 根据传入的Type的顺序orders，构建最终的SQL
	for _, order := range orders {
		// 按顺序取出构建好的子句，拼接
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
