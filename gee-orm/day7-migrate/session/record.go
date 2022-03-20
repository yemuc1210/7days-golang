package session

import (
	"errors"
	"geeorm/clause"
	"reflect"
)

// Insert one or more records in database
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		s.CallMethod(BeforeInsert, value)
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
	// 钩子函数  	s.CallMethod(BeforeQuery, nil)
	s.CallMethod(BeforeQuery, nil)
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
		// 钩子函数     `AfterQuery` 钩子可以操作每一行记录。
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		// reflect.Set方法
		// 将dest 添加到切片放在
		destSlice.Set(reflect.Append(destSlice, dest))
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

// First gets the 1st row
// 只返回一个结构
// 调用形式：
// u := &User{}
// _ = s.OrderBy("Age DESC").First(u)
func (s *Session) First(value interface{}) error {
	// 对象类型
	dest := reflect.Indirect(reflect.ValueOf(value))
	// SliceOf 返回一种类型的切片
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}

// Limit adds limit condition to clause
// 添加limit的链调用方法
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// Where adds limit condition to clause
// where 条件查询子句适合链式调用
func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

// OrderBy adds order by condition to clause
// OrderBy 条件查询子句适合链式调用
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

/// 使用where子句进行更新
// 支持字典类型参数
// 或者是 kv 形式: "Name", "Tom", "Age", 18, ....
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	/// 判断一下传入参数的类型
	m, ok := kv[0].(map[string]interface{})
	// 不是map类型
	if !ok {
		// 将传入参数构建为mao类型
		m = make(map[string]interface{})
		// 两两一对
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	// 调用_update构建update子句
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

// Delete records with where clause
func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	// 调用_delete生成子句，知晓表名即可
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

// Count records with where clause
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}
