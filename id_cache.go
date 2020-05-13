package data

import (
	"github.com/qingfenghuohu/tools/redis"
	"strings"
)

type IdReal struct {
	dck []DataCacheKey
}

func (real *IdReal) SetCacheData(rcd []RealCacheData) {
	CacheData := map[string][]interface{}{}
	Keys := map[string]map[int64][]string{}
	for _, v := range rcd {
		if len(CacheData[v.CacheKey.ConfigName]) == 0 {
			CacheData[v.CacheKey.ConfigName] = []interface{}{}
			Keys[v.CacheKey.ConfigName] = map[int64][]string{}
		}
		if len(Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)]) == 0 {
			Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)] = []string{}
		}
		CacheData[v.CacheKey.ConfigName] = append(CacheData[v.CacheKey.ConfigName], v.CacheKey.String())
		CacheData[v.CacheKey.ConfigName] = append(CacheData[v.CacheKey.ConfigName], v.Result)
		Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)] = append(Keys[v.CacheKey.ConfigName][int64(v.CacheKey.LifeTime)], v.CacheKey.String())
	}
	for key, val := range CacheData {
		redis.GetInstance(key).MSet(val...)
		for k, v := range Keys[key] {
			redis.GetInstance(key).Expire(k, v...)
		}
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
	var result []RealCacheData
	realData := make(map[string]DataCacheKey)
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
		param := strings.Join(realParams[i], ",")
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
			result = append(result, RealCacheData{CacheKey: v, Result: vv})
		}
	}
	return result
}
func (real *IdReal) SetDataCacheKey(dck []DataCacheKey) {
	real.dck = RemoveDuplicateElement(dck)
}
func (real *IdReal) DelCacheData() {
	keys := map[string][]interface{}{}
	for _, v := range real.dck {
		if len(keys[v.ConfigName]) == 0 {
			keys[v.ConfigName] = []interface{}{}
		}
		keys[v.ConfigName] = append(keys[v.ConfigName], v.String())
	}
	for key, val := range keys {
		redis.GetInstance(key).Delete(val...)
	}
}
