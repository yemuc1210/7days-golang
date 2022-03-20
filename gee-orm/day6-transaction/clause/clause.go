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

// 定义Clause支持的类型
// 对应子句的generator，定义在generators中
const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	// UPDATE 新增加三个类型
	// 实现更新
	UPDATE
	// DELETE 删除
	DELETE
	// COUNT 统计
	COUNT
)

// Set adds a sub clause of specific type
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// Build generate the final SQL and SQLVars
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
