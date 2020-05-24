package dao

import (
	"encoding/json"
	"github.com/qingfenghuohu/tools/str"
	"sync"
)

type Result struct {
	data      sync.Map
	WaitGroup sync.WaitGroup
}
type resultInfo struct {
	data interface{}
}

func (r *Result) add() {
	r.WaitGroup.Add(1)
}

func (r *Result) done() {
	r.WaitGroup.Done()
}

func (r *Result) wait() {
	r.WaitGroup.Wait()
}

func (r *Result) write(key string, val interface{}) {
	r.data.Store(key, val)
}

func (r *Result) read(key string) interface{} {
	result, _ := r.data.Load(key)
	return result
}

func (r *Result) exist(key string) bool {
	_, result := r.data.Load(key)
	return result
}

func (r *Result) del(key string) {
	r.data.Delete(key)
}

func (r *Result) Read(model ModelInfo, key string, params ...string) resultInfo {
	var result resultInfo
	k := CreateCacheKey(model, key, params...)
	res := []map[string]interface{}{}
	tmp := r.read(k.String())
	if tmp != nil && typeof(tmp) != "[]map[string]interface {}" {
		json.Unmarshal([]byte(tmp.(string)), &res)
	}
	result.data = res
	return result
}

func (r resultInfo) SliceMap() []map[string]string {
	result := []map[string]string{}
	for _, val := range r.data.([]map[string]interface{}) {
		tmp := map[string]string{}
		for k, v := range val {
			tmp[k] = str.Obj2Str(v)
		}
		result = append(result, tmp)
	}
	return result
}

func (r resultInfo) Map() map[string]string {
	result := map[string]string{}
	data := r.data.([]map[string]interface{})
	if len(data) == 0 {
		return result
	}
	for key, val := range data[0] {
		result[key] = str.Obj2Str(val)
	}
	return result
}

func (r resultInfo) String() string {
	return r.data.(string)
}

func (r resultInfo) Int() int {
	return r.data.(int)
}
func (r resultInfo) Bool() bool {
	return r.data.(bool)
}
