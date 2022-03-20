package session

import (
	"geeorm/log"
	"reflect"
)
// 钩子： 在可能增加功能地方预设构造，在需要重新修改或添加逻辑的时候，把扩展的类或者方法挂载到该点
// - 在可能发生变化的地方
// - 感觉有点像触发器。当某个条件发生时，会触发固定的操作。

// 钩子个结构体绑定，每个结构体需要首先各自的钩子

// Hooks 常量
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

// CallMethod 调用注册的钩子，注册
func (s *Session) CallMethod(method string, value interface{}) {
	// s.RefTable() 获得session s的 refTable *schema.Schema
	// 获得对应的模型名称，也就得到了实例，然后调用方法
	// 当前会话正在操作的对象
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	// 参数   将session作为入参，每一个钩子的入参类型都是*Session
	param := []reflect.Value{reflect.ValueOf(s)}
	// fm 是Value 则 true
	if fm.IsValid() {
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
	return
}
