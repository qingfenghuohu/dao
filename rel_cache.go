package dao

import (
	"github.com/qingfenghuohu/tools/redis"
	"strconv"
)

type RelReal struct {
	dck []CacheKey
}

func (real *RelReal) SetCacheData(rcd []RealCacheData) {
	Keys := map[string][]redis.HMSMD{}
	for _, v := range rcd {
		Hmsmd := redis.HMSMD{}
		if len(Keys[v.CacheKey.ConfigName]) == 0 {
			Keys[v.CacheKey.ConfigName] = []redis.HMSMD{}
		}
		Hmsmd.Key = v.CacheKey.GetCacheKey()
		Hmsmd.Data = map[string]interface{}{v.CacheKey.Params[1]: v.Result}
		Hmsmd.Ttl = v.CacheKey.LifeTime
		Keys[v.CacheKey.ConfigName] = append(Keys[v.CacheKey.ConfigName], Hmsmd)
	}
	for key, val := range Keys {
		redis.GetInstance(key).HMSetMulti(val)
	}
}
func (real *RelReal) GetCacheData(res *Result) {
	Keys := map[string]map[string][]string{}
	for _, v := range real.dck {
		if len(Keys[v.ConfigName]) == 0 {
			Keys[v.ConfigName] = map[string][]string{}
		}
		if len(Keys[v.ConfigName][v.GetCacheKey()]) == 0 {
			Keys[v.ConfigName][v.GetCacheKey()] = []string{}
		}
		if v.Params[1] != "" {
			Keys[v.ConfigName][v.GetCacheKey()] = append(Keys[v.ConfigName][v.GetCacheKey()], v.Params[1])
		}
	}
	GetKeys := map[string]int{}
	GetField := map[string]int{}
	for k, v := range Keys {
		tmp := redis.GetInstance(k).HMGetMulti(v)
		for key, val := range tmp {
			GetKeys[key] = 1
			for kk, vv := range val {
				GetField[key+"_"+kk] = 1
				res.write(key+"_"+kk, vv)
			}
		}
	}
	for _, v := range real.dck {
		if GetKeys[v.GetCacheKey()] == 1 && GetField[v.String()] != 1 {
			res.write(v.String(), v.DefaultVal)
		}
	}
}
func (real *RelReal) GetRealData() []RealCacheData {
	var result RealData
	dataCacheKey := map[string]map[string][]CacheKey{}
	models := map[string]ModelInfo{}
	for _, v := range real.dck {
		TableName := v.Model.MicroName() + "." + v.Model.DbName() + "." + v.Model.TableName()
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
	result.Wait()
	for _, v := range real.dck {
		v.Params = []string{"", ""}
		result.Append(RealCacheData{CacheKey: v, Result: ""})
	}
	return result.Data
}
func (real *RelReal) SetDataCacheKey(dck []CacheKey) Cache {
	real.dck = RemoveDuplicateCacheKey(dck)
	return real
}
func (real *RelReal) DelCacheData(dck []CacheKey) {
	keys := map[string][]CacheKey{}
	for _, v := range dck {
		if len(keys[v.ConfigName]) == 0 {
			keys[v.ConfigName] = []CacheKey{}
		}
		keys[v.ConfigName] = append(keys[v.ConfigName], v)
	}
	for key, val := range keys {
		data := []interface{}{}
		for _, v := range val {
			if v.Params[1] != "" {
				data = append(data, v.GetCacheKey())
				data = append(data, v.Params[1])
			}
		}
		if len(data) > 0 {
			redis.GetInstance(key).HMDel(data)
		}
	}
}
func (real *RelReal) GetCacheKey(key *CacheKey) string {
	var result string
	result = strconv.Itoa(key.Version) + ":" +
		key.CType + ":" +
		key.Model.MicroName() + "." + key.Model.DbName() + "." + key.Model.TableName() + ":" +
		key.Key + ":"
	if key.Params[0] != "" {
		result += key.Params[0]
	}
	return result
}
func (real *RelReal) DbToCache(md *ModelData) []RealCacheData {
	var result []RealCacheData
	mddb := md.Model.DbToCache(md, CacheTypeRelation)
	if len(mddb.DelData) > 0 {
		real.DelCacheData(RemoveDuplicateCacheKey(mddb.DelData))
	}
	if len(mddb.SaveData) > 0 {
		tmp := real.SetDataCacheKey(RemoveDuplicateCacheKey(mddb.SaveData)).GetRealData()
		result = append(result, tmp...)
	}
	if len(mddb.Data) > 0 {
		result = append(result, mddb.Data...)
	}
	return result
}
