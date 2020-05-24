package dao

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CacheTypeIds      = "id"
	CacheTypeRelation = "rel"
	CacheTypeI        = "i"
	CacheTypeField    = "f"
	CacheTypeTotal    = "t"
)

type Cache interface {
	SetCacheData(rcd []RealCacheData)
	GetCacheData(res *Result)
	GetRealData() []RealCacheData
	SetDataCacheKey(dck []CacheKey) Cache
	DelCacheData(dck []CacheKey)
	DbToCache(md ModelData) []RealCacheData
	GetCacheKey(key *CacheKey) string
}

type CacheKey struct {
	Key        string
	CType      string
	Model      ModelInfo
	Params     []string
	LifeTime   int
	ResetTime  int
	ResetCount int
	Version    int
	RelField   []string
	ResField   []string
	ResetType  int //0=重建,1=删除
	ConfigName string
	Operation  string
	DefaultVal interface{}
}

type ListCacheKey struct {
	data []CacheKey
}

type RealCacheData struct {
	Result   interface{}
	CacheKey CacheKey
}
type RealData struct {
	Data      []RealCacheData
	SaveData  []CacheKey
	DelData   []CacheKey
	Sync      sync.Mutex
	WaitGroup sync.WaitGroup
}
type TypeRealData struct {
	Data      map[string][]RealCacheData
	Sync      sync.Mutex
	WaitGroup sync.WaitGroup
}

func ReBuild(resetKey []CacheKey) []RealCacheData {
	t := time.Now()
	var result RealData
	typeResetKey := GetTypeDataCacheKey(resetKey)
	//获取真实数据
	for key, val := range typeResetKey {
		result.Add()
		go func(k string, v []CacheKey, result *RealData) {
			rc := RunCache(k)
			rc.SetDataCacheKey(v)
			RealCacheData := rc.GetRealData()
			rc.SetCacheData(RealCacheData)
			result.Append(RealCacheData...)
			result.Done()
		}(key, val, &result)
	}
	result.Wait()
	elapsed := time.Since(t)
	fmt.Println("ReBuild:", elapsed)
	return result.Data
}

func getCache(configKey []CacheKey) (Result, []CacheKey) {
	var result Result
	var defect []CacheKey
	for i, _ := range configKey {
		configKey[i].SetOperation("get")
	}
	dataCacheKey := GetTypeDataCacheKey(configKey)
	for key, val := range dataCacheKey {
		result.add()
		go func(k string, v []CacheKey, result *Result) {
			rc := RunCache(k)
			rc.SetDataCacheKey(v)
			rc.GetCacheData(result)
			result.done()
		}(key, val, &result)
	}
	result.wait()
	for _, v := range configKey {
		if result.read(v.String()) == nil {
			defect = append(defect, v)
		}
	}
	return result, defect
}

func RunCache(key string) Cache {
	var result Cache
	switch key {
	case CacheTypeRelation:
		result = &RelReal{}
	case CacheTypeIds:
		result = &IdReal{}
	case CacheTypeField:
		result = &FieldReal{}
	case CacheTypeTotal:
		result = &TotalReal{}
	case CacheTypeI:
		result = &IReal{}
	default:
		result = &Real{}
	}
	return result
}

func CreateCacheKey(m ModelInfo, key string, p ...string) CacheKey {
	result := Model(m).modelInfo.GetDataCacheKey()[key]
	result.Params = p
	return result
}

func GetData(configKey []CacheKey) Result {
	t := time.Now()
	configKey = RemoveDuplicateCacheKey(configKey)
	//获取全部缓存数据
	AllData, resetKey := getCache(configKey)
	//重置缓存数据
	if len(resetKey) > 0 {
		RealCacheData := ReBuild(resetKey)
		for _, v := range RealCacheData {
			AllData.write(v.CacheKey.String(), v.Result)
		}
	}
	elapsed := time.Since(t)
	fmt.Println("GetData:", elapsed)
	return AllData
}

func SaveCache(md ModelData) {
	data := DbToTypeCache(md)
	waitGroup := sync.WaitGroup{}
	for k, v := range data {
		waitGroup.Add(1)
		go func(key string, val []RealCacheData, w *sync.WaitGroup) {
			RunCache(key).SetCacheData(val)
			w.Done()
		}(k, v, &waitGroup)
	}
	waitGroup.Wait()
}

func DbToTypeCache(md ModelData) map[string][]RealCacheData {
	result := TypeRealData{}
	dataCacheKey := md.Model.GetDataCacheKey()
	for _, confVal := range dataCacheKey {
		result.Add()
		go func(confVal CacheKey, result *TypeRealData, md ModelData) {
			tmp := RunCache(confVal.CType).DbToCache(md)
			result.append(confVal.CType, tmp...)
			result.Done()
		}(confVal, &result, md)
	}
	result.Wait()
	return result.Data
}

func (dck *CacheKey) String() string {
	return strconv.Itoa(dck.Version) + ":" +
		dck.CType + ":" +
		dck.Model.MicroName() + "." + dck.Model.DbName() + "." + dck.Model.TableName() + ":" +
		dck.Key + ":" +
		strings.Join(dck.Params, "_")
}
func (dck *CacheKey) GetCacheKey() string {
	return RunCache(dck.CType).GetCacheKey(dck)
}
func (dck *CacheKey) SetOperation(val string) {
	dck.Operation = strings.ToLower(val)
}

func (rd *RealData) Append(data ...RealCacheData) {
	rd.Sync.Lock()
	rd.Data = append(rd.Data, data...)
	rd.Sync.Unlock()
}

func (rd *RealData) Add() {
	rd.WaitGroup.Add(1)
}

func (rd *RealData) Done() {
	rd.WaitGroup.Done()
}

func (rd *RealData) Wait() {
	rd.WaitGroup.Wait()
}

func (rd *RealData) SaveAppend(data ...CacheKey) {
	rd.Sync.Lock()
	rd.SaveData = append(rd.SaveData, data...)
	rd.Sync.Unlock()
}

func (rd *RealData) DelAppend(data ...CacheKey) {
	rd.Sync.Lock()
	rd.DelData = append(rd.DelData, data...)
	rd.Sync.Unlock()
}

func (rd *TypeRealData) append(key string, data ...RealCacheData) {
	rd.Sync.Lock()
	if len(rd.Data[key]) == 0 {
		rd.Data[key] = []RealCacheData{}
	}
	rd.Data[key] = append(rd.Data[key], data...)
	rd.Sync.Unlock()
}
func (rd *TypeRealData) Add() {
	rd.WaitGroup.Add(1)
}

func (rd *TypeRealData) Done() {
	rd.WaitGroup.Done()
}

func (rd *TypeRealData) Wait() {
	rd.WaitGroup.Wait()
}

func GetCache() *ListCacheKey {
	result := ListCacheKey{}
	return &result
}

func (ld *ListCacheKey) Add(model ModelInfo, key string, params ...string) *ListCacheKey {
	ld.data = append(ld.data, CreateCacheKey(model, key, params...))
	return ld
}

func (ld *ListCacheKey) GetData() Result {
	return GetData(ld.data)
}

func (ld *ListCacheKey) typeCacheKey(key string, m ModelInfo) []CacheKey {
	result := map[string][]CacheKey{}
	data := []CacheKey{}
	tmp := m.GetDataCacheKey()
	for _, v := range tmp {
		data = append(data, v)
	}
	for _, v := range data {
		if len(result[v.CType]) == 0 {
			result[v.CType] = []CacheKey{}
		}
		result[v.CType] = append(result[v.CType], v)
	}
	return result[key]
}
func (ld *ListCacheKey) operationCacheKey(data []CacheKey) map[string][]CacheKey {
	result := map[string][]CacheKey{}
	for _, v := range data {
		if v.Operation == "" {
			v.SetOperation("get")
		}
		if len(result[v.Operation]) == 0 {
			result[v.Operation] = []CacheKey{}
		}
		result[v.Operation] = append(result[v.Operation], v)
	}
	return result
}
func resField(field []string, data interface{}) []map[string]interface{} {
	result := []map[string]interface{}{}
	d := data.([]map[string]interface{})
	if len(field) == 0 {
		return d
	}
	for _, val := range d {
		tmp := map[string]interface{}{}
		for _, v := range field {
			if val[v] != nil {
				tmp[v] = val[v]
			}
		}
		result = append(result, tmp)
	}
	return result
}
func setDefaultRealCacheData(rcd []CacheKey, result []RealCacheData) []RealCacheData {
	for _, val := range rcd {
		i := 1
		for _, v := range result {
			if val.String() == v.CacheKey.String() && i == 1 {
				i = 2
				break
			}
		}
		if i == 1 {
			result = append(result, RealCacheData{[]map[string]interface{}{}, val})
		}
	}
	return result
}
