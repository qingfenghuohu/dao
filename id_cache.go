package dao

import (
	"fmt"
	"github.com/qingfenghuohu/tools"
	"github.com/qingfenghuohu/tools/redis"
	"strings"
	"time"
)

type IdReal struct {
	dck []CacheKey
}

func (real *IdReal) SetCacheData(rcd []RealCacheData) {
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
func (real *IdReal) GetCacheData(res *Result) {
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
func (real *IdReal) GetRealData() []RealCacheData {
	t := time.Now()
	var result []RealCacheData
	realData := make(map[string]CacheKey)
	realParams := make(map[string][]string)
	for _, v := range real.dck {
		tmpKey := v.Model.DbName() + "_" + v.Model.TableName()
		if ok := realData[tmpKey]; ok.CType == "" {
			realData[tmpKey] = v
			realParams[tmpKey] = []string{}
		}
		if v.Params[0] != "" {
			realParams[tmpKey] = append(realParams[tmpKey], v.Params[0])
		}
	}
	for i, v := range realData {
		tmpIds := tools.RemoveDuplicateElement(realParams[i])
		param := strings.Join(tmpIds, ",")
		m := Model(v.Model).InitField()
		if m.pk == "" {
			m.pk = "Id"
		}
		if m.pkMysql == "" {
			m.pkMysql = "id"
		}
		res := Model(v.Model).Where(m.pkMysql + " in(" + param + ")").Select()
		for _, vv := range res {
			v.Params = []string{vv[m.pk].(string)}
			res := []map[string]interface{}{}
			res = append(res, vv)
			result = append(result, RealCacheData{CacheKey: v, Result: res})
		}
	}
	result = setDefaultRealCacheData(real.dck, result)
	elapsed := time.Since(t)
	fmt.Println("GetRealData:", elapsed)
	return result
}
func (real *IdReal) SetDataCacheKey(dck []CacheKey) Cache {
	real.dck = RemoveDuplicateCacheKey(dck)
	return real
}
func (real *IdReal) GetCacheKey(key *CacheKey) string {
	return key.String()
}
func (real *IdReal) DelCacheData(dck []CacheKey) {
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
func (real *IdReal) DbToCache(md *ModelData) []RealCacheData {
	var result RealData
	pk := Model(md.Model).InitField().pk
	for _, val := range md.Data {
		if md.Operation == "del" {
			val.AfterData = val.BeData
		}
		res := []map[string]interface{}{}
		res = append(res, val.AfterData)
		id := val.AfterData[pk].(string)
		rcd := RealCacheData{}
		rcd.CacheKey = CreateCacheKey(md.Model, CacheTypeIds, id)
		rcd.CacheKey.SetOperation(md.Operation)
		rcd.Result = res
		result.Data = append(result.Data, rcd)
	}

	return result.Data
}
