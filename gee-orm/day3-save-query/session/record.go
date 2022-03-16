package session

import (
	"geeorm/clause"
	"reflect"
)

// 记录增删查改相关代码

// Insert one or more records in database
// 插入数据库操作
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		// values 是若干个对象实例
		// 调用Model(value)将对象实例解析为数据库表
		// RefTable() 返回一个schema实例，也就是一个数据库表实例
		table := s.Model(value).RefTable()
		// insert into tableName (%s)
		// FieldNames 获取字段名
		// Set生成子句，SQL存放在Clause对象放在
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		// 值通过RecordValues调整顺序，和数据库表的字段顺序一致
		recordValues = append(recordValues, table.RecordValues(value))
	}

	//  拼接Values (),()....
	s.clause.Set(clause.VALUES, recordValues...)
	// 根据传入的类型顺序，构建最终的SQL语句
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find gets all eligible records
// 期望的效果是：传入一个切片指针，将查询结果保存在切片中
//s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
//var users []User
//s.Find(&users)
// 因此Find需要根据平铺的字段值构造出对象
func (s *Session) Find(values interface{}) error {
	// 获得切片实例
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	// 得到切片类型
	destType := destSlice.Type().Elem()
	// Model根据对象，构建schema实例，也就是表，然后用RefTable返回得到这个表实例
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	// 构建查询子句
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	// 得到查询结果
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		// 根据FieldName，构建字段名称的切片。这里，对象实例的名称和数据库表的列名是一致
		for _, name := range table.FieldNames {
			// 由于Scan参数一般是指针，所以有调用Addr()，返回指针
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// 调用Scan方法，将改行记录每一列的值，依次付给values中的每一个字段
		if err := rows.Scan(values...); err != nil {
			return err
		}
		// reflect.Set方法
		// 将dest 添加到切片放在
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}
