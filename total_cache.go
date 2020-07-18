package dao

import (
	"github.com/qingfenghuohu/tools"
	"github.com/qingfenghuohu/tools/redis"
	"strconv"
)

type TotalReal struct {
	dck []CacheKey
}

//Operation = get,plus,reduce
func (real *TotalReal) SetCacheData(rcd []RealCacheData) {
	CacheData := map[string][]interface{}{}
	Keys := map[string]map[int64][]string{}
	PlusData := map[string]map[string]int{}
	ReduceData := map[string]map[string]int{}
	SetData := map[string]map[string]int{}
	DelData := []CacheKey{}
	ck := []CacheKey{}
	var dbresult int
	for _, v := range rcd {
		ck = append(ck, v.CacheKey)
		dbresult, _ = strconv.Atoi(tools.Obj2Str(v.Result))
		if v.CacheKey.Operation == "del" {
			DelData = append(DelData, v.CacheKey)
		} else if v.CacheKey.Operation == "plus" {
			if len(PlusData[v.CacheKey.ConfigName]) == 0 {
				PlusData[v.CacheKey.ConfigName] = map[string]int{}
			}
			if _, ok := PlusData[v.CacheKey.ConfigName][v.CacheKey.String()]; ok {
				PlusData[v.CacheKey.ConfigName][v.CacheKey.String()] += dbresult
			} else {
				PlusData[v.CacheKey.ConfigName][v.CacheKey.String()] = dbresult
			}
		} else if v.CacheKey.Operation == "reduce" {
			if len(ReduceData[v.CacheKey.ConfigName]) == 0 {
				ReduceData[v.CacheKey.ConfigName] = map[string]int{}
			}
			if _, ok := ReduceData[v.CacheKey.ConfigName][v.CacheKey.String()]; ok {
				ReduceData[v.CacheKey.ConfigName][v.CacheKey.String()] += dbresult
			} else {
				ReduceData[v.CacheKey.ConfigName][v.CacheKey.String()] = dbresult
			}
		} else {
			if len(CacheData[v.CacheKey.ConfigName]) == 0 {
				CacheData[v.CacheKey.ConfigName] = []interface{}{}
				Keys[v.CacheKey.ConfigName] = map[int64][]string{}
			}
			if len(Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)]) == 0 {
				Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)] = []string{}
			}
			if len(SetData[v.CacheKey.ConfigName]) == 0 {
				SetData[v.CacheKey.ConfigName] = map[string]int{}
			}
			CacheData[v.CacheKey.ConfigName] = append(CacheData[v.CacheKey.ConfigName], v.CacheKey.String())
			CacheData[v.CacheKey.ConfigName] = append(CacheData[v.CacheKey.ConfigName], dbresult)
			SetData[v.CacheKey.ConfigName][v.CacheKey.String()] = dbresult
			Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)] = append(Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)], v.CacheKey.String())
		}

	}
	if len(CacheData) > 0 {
		for key, val := range CacheData {
			redis.GetInstance(key).MSet(val...)
			for k, v := range Keys[key] {
				redis.GetInstance(key).Expire(k, v...)
			}
		}
	}
	_, falseCk := ExistsMulti(ck)
	if len(PlusData) > 0 {
		for key, val := range PlusData {
			for _, v := range falseCk {
				delete(val, v.String())
			}
			if len(val) > 0 {
				redis.GetInstance(key).IncrByMulti(val)
			}
		}
	}
	if len(ReduceData) > 0 {
		for key, val := range ReduceData {
			for _, v := range falseCk {
				delete(val, v.String())
			}
			if len(val) > 0 {
				redis.GetInstance(key).DecrByMulti(val)
			}
		}
	}
	if len(DelData) > 0 {
		real.DelCacheData(DelData)
	}
}
func (real *TotalReal) GetCacheData(res *Result) {
	Keys := map[string][]string{}
	for _, v := range real.dck {
		if len(Keys[v.ConfigName]) == 0 {
			Keys[v.ConfigName] = []string{}
		}
		Keys[v.ConfigName] = append(Keys[v.ConfigName], v.String())
	}
	for k, v := range Keys {
		tmp := redis.GetInstance(k).MGet(v...)
		for key, val := range tmp {
			res.write(key, val)
		}
	}
}
func (real *TotalReal) GetRealData() []RealCacheData {
	var result RealData
	dataCacheKey := map[string]map[string][]CacheKey{}
	models := map[string]ModelInfo{}
	for _, v := range real.dck {
		TableName := v.Model.DbName() + "." + v.Model.TableName()
		if len(dataCacheKey[TableName]) == 0 {
			dataCacheKey[TableName] = map[string][]CacheKey{}
		}
		if len(dataCacheKey[TableName][v.Key]) == 0 {
			dataCacheKey[TableName][v.Key] = []CacheKey{}
		}
		dataCacheKey[TableName][v.Key] = append(dataCacheKey[TableName][v.Key], v)
		models[TableName] = v.Model
	}
	for key, val := range dataCacheKey {
		result.Add()
		go func(key string, val map[string][]CacheKey, result *RealData) {
			tmp := models[key].GetRealData(val)
			result.Append(tmp...)
			result.Done()
		}(key, val, &result)
	}
	for key, val := range dataCacheKey {
		tmp := models[key].GetRealData(val)
		result.Append(tmp...)
	}
	return result.Data
}
func (real *TotalReal) SetDataCacheKey(dck []CacheKey) Cache {
	real.dck = RemoveDuplicateCacheKey(dck)
	return real
}
func (real *TotalReal) GetCacheKey(key *CacheKey) string {
	return key.String()
}
func (real *TotalReal) DelCacheData(dck []CacheKey) {
	keys := map[string][]string{}
	for _, v := range dck {
		if len(keys[v.ConfigName]) == 0 {
			keys[v.ConfigName] = []string{}
		}
		keys[v.ConfigName] = append(keys[v.ConfigName], v.String())
	}
	for key, val := range keys {
		redis.GetInstance(key).Delete(val...)
	}
}
func (real *TotalReal) DbToCache(md *ModelData) []RealCacheData {
	var result []RealCacheData
	mddb := md.Model.DbToCache(md, CacheTypeTotal)
	if len(mddb.SaveData) > 0 {
		tmp := real.SetDataCacheKey(RemoveDuplicateCacheKey(mddb.SaveData)).GetRealData()
		result = append(result, tmp...)
	}
	if len(mddb.DelData) > 0 {
		real.DelCacheData(RemoveDuplicateCacheKey(mddb.DelData))
	}
	if len(mddb.Data) > 0 {
		result = append(result, mddb.Data...)
	}
	return result
}
