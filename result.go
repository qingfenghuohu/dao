package dao

import (
	"encoding/json"
	"fmt"
	"github.com/qingfenghuohu/tools/str"
	"strconv"
	"sync"
)

type Result struct {
	data      map[string]interface{}
	WaitGroup sync.WaitGroup
	Sync      sync.Mutex
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
func (r *Result) Range() {
	for k, v := range r.data {
		fmt.Println(k, v)
	}
}
func (r *Result) write(key string, val interface{}) {
	r.Sync.Lock()
	if len(r.data) <= 0 {
		r.data = map[string]interface{}{}
	}
	r.data[key] = val
	r.Sync.Unlock()
}

func (r *Result) read(key string) interface{} {
	result, _ := r.data[key]
	return result
}

func (r *Result) exist(key string) bool {
	_, result := r.data[key]
	return result
}

func (r *Result) del(key string) {
	r.Sync.Lock()
	if len(r.data) <= 0 {
		r.data = map[string]interface{}{}
	} else {
		delete(r.data, key)
	}
	r.Sync.Unlock()
}

func (r *Result) Read(model ModelInfo, key string, params ...string) resultInfo {
	result := resultInfo{}
	k := CreateCacheKey(model, key, params...)
	res := []map[string]interface{}{}
	tmp := r.read(k.String())
	if k.CType == CacheTypeRelation || k.CType == CacheTypeTotal || typeof(tmp) == "[]map[string]interface {}" {
		result.data = tmp
	} else {
		if tmp != nil && typeof(tmp) != "[]map[string]interface {}" {
			json.Unmarshal([]byte(tmp.(string)), &res)
		}
		result.data = res
	}
	return result
}

func (r *Result) Exist(model ModelInfo, key string, params ...string) bool {
	k := CreateCacheKey(model, key, params...)
	_, result := r.data[k.String()]
	return result
}

func (r resultInfo) SliceMap() []map[string]string {
	result := []map[string]string{}
	for _, val := range r.data.([]map[string]interface{}) {
		tmp := map[string]string{}
		for k, v := range val {
			tmp[k] = r.getVal(v)
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
		result[key] = r.getVal(val)
	}
	return result
}

func (r resultInfo) String() string {
	return r.getVal(r.data)
}

func (r resultInfo) Int() int {
	var result int
	if typeof(r.data) == "string" {
		result, _ = strconv.Atoi(r.data.(string))
	} else {
		result = r.data.(int)
	}
	return result
}
func (r resultInfo) Bool() bool {
	return r.data.(bool)
}
func (r resultInfo) Raw() interface{} {
	return r.data
}

func (r resultInfo) getVal(data interface{}) string {
	result := str.Obj2Str(data)
	if result == "NULL" || result == "null" {
		result = ""
	}
	return result
}
