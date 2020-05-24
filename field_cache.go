package dao

import (
	"github.com/qingfenghuohu/tools/redis"
	"github.com/qingfenghuohu/tools/str"
	"strings"
)

type FieldReal struct {
	dck []CacheKey
}

func (real *FieldReal) SetCacheData(rcd []RealCacheData) {
	CacheData := map[string][]interface{}{}
	Keys := map[string]map[int64][]string{}
	delCacheKey := []CacheKey{}
	for _, v := range rcd {
		if v.CacheKey.Operation == "del" {
			delCacheKey = append(delCacheKey, v.CacheKey)
		} else {
			if v.CacheKey.ResetType == 0 {
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
func (real *FieldReal) GetCacheData(res *Result) {
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
func (real *FieldReal) GetRealData() []RealCacheData {
	var result RealData
	for _, val := range real.dck {
		if len(val.RelField) == 0 {
			continue
		}
		result.Add()
		go func(val CacheKey, result *RealData) {
			res := RealCacheData{}
			tmp := []string{}
			tmp1 := []interface{}{}
			for i, v := range val.RelField {
				if val.Params[i] != "" {
					tmp = append(tmp, v+" = ?")
					tmp1 = append(tmp1, val.Params[i])
				}
			}
			if len(tmp) > 0 && len(tmp1) > 0 {
				sql := strings.Join(tmp, " and ")
				res.CacheKey = val
				res.Result = Model(val.Model).Where(sql, tmp1...).Select()
				result.Append(res)
			}
			result.Done()
		}(val, &result)
	}
	result.Wait()
	result.Data = setDefaultRealCacheData(real.dck, result.Data)
	return result.Data
}
func (real *FieldReal) SetDataCacheKey(dck []CacheKey) Cache {
	real.dck = RemoveDuplicateCacheKey(dck)
	return real
}
func (real *FieldReal) GetCacheKey(key *CacheKey) string {
	return key.String()
}
func (real *FieldReal) DelCacheData(dck []CacheKey) {
	keys := map[string][]interface{}{}
	for _, v := range dck {
		if len(keys[v.ConfigName]) == 0 {
			keys[v.ConfigName] = []interface{}{}
		}
		keys[v.ConfigName] = append(keys[v.ConfigName], v.String())
	}
	for key, val := range keys {
		redis.GetInstance(key).Delete(val...)
	}
}
func (real *FieldReal) DbToCache(md *ModelData) []RealCacheData {
	var result RealData
	rcd := GetCache().typeCacheKey(CacheTypeField, md.Model)
	for _, val := range md.Data {
		for _, v := range rcd {
			for _, vv := range v.RelField {
				value := str.Obj2Str(val.BeData[vv])
				v.Params = append(v.Params, value)
			}
			be := v
			v.Params = []string{}
			for _, vv := range v.RelField {
				value := str.Obj2Str(val.AfterData[vv])
				v.Params = append(v.Params, value)
			}
			after := v
			if md.Operation == "del" || v.ResetType == 0 {
				result.DelData = append(result.DelData, be)
			} else {
				result.SaveData = append(result.SaveData, after)
			}
		}
	}
	result.SaveData = RemoveDuplicateCacheKey(result.SaveData)
	result.DelData = RemoveDuplicateCacheKey(result.DelData)
	real.DelCacheData(result.DelData)
	return real.SetDataCacheKey(result.SaveData).GetRealData()
}
