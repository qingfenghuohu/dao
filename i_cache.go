package dao

import "github.com/qingfenghuohu/tools/redis"

type IReal struct {
	dck []CacheKey
}

func (real *IReal) SetCacheData(rcd []RealCacheData) {
	CacheData := map[string][]interface{}{}
	Keys := map[string]map[int64][]string{}
	delCacheKey := []CacheKey{}
	for _, v := range rcd {
		if v.CacheKey.Operation == "del" || v.CacheKey.ResetType == 0 {
			delCacheKey = append(delCacheKey, v.CacheKey)
		} else {
			if len(CacheData[v.CacheKey.ConfigName]) == 0 {
				CacheData[v.CacheKey.ConfigName] = []interface{}{}
				Keys[v.CacheKey.ConfigName] = map[int64][]string{}
			}
			if len(Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)]) == 0 {
				Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)] = []string{}
			}
			CacheData[v.CacheKey.ConfigName] = append(CacheData[v.CacheKey.ConfigName], v.CacheKey.String())
			CacheData[v.CacheKey.ConfigName] = append(CacheData[v.CacheKey.ConfigName], resField(v.CacheKey.ResField, v.Result))
			Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)] = append(Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)], v.CacheKey.String())
		}
	}
	if len(CacheData) > 0 {
		for key, val := range CacheData {
			redis.GetInstance(key).MSetJson(val...)
			for k, v := range Keys[key] {
				redis.GetInstance(key).Expire(k, v...)
			}
		}
	}
	if len(delCacheKey) > 0 {
		real.DelCacheData(delCacheKey)
	}
}
func (real *IReal) GetCacheData(res *Result) {
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
func (real *IReal) GetRealData() []RealCacheData {
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
	return result.Data
}
func (real *IReal) SetDataCacheKey(dck []CacheKey) Cache {
	real.dck = RemoveDuplicateCacheKey(dck)
	return real
}
func (real *IReal) GetCacheKey(key *CacheKey) string {
	return key.String()
}
func (real *IReal) DelCacheData(dck []CacheKey) {
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
func (real *IReal) DbToCache(md *ModelData) []RealCacheData {
	var result []RealCacheData
	mddb := md.Model.DbToCache(md, CacheTypeI)
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
