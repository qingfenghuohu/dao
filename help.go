package dao

import (
	"errors"
	"github.com/qingfenghuohu/tools"
	"github.com/qingfenghuohu/tools/redis"
	"strconv"
	"time"
)

func RemoveDuplicateCacheKey(data []CacheKey) []CacheKey {
	var result []CacheKey
	for _, v := range data {
		if len(result) == 0 {
			result = append(result, v)
		} else {
			for _, val := range result {
				if v.String() != val.String() {
					result = append(result, v)
				}
			}
		}
	}
	return result
}

func GetTypeDataCacheKey(data []CacheKey) map[string][]CacheKey {
	result := map[string][]CacheKey{}
	for _, v := range data {
		if ok := result[v.CType]; len(ok) == 0 {
			result[v.CType] = []CacheKey{}
		}
		result[v.CType] = append(result[v.CType], v)
	}
	return result
}
func Exists(ck CacheKey) bool {
	var result bool
	var key string
	if ck.CType == CacheTypeRelation {
		key = ck.GetCacheKey()
	} else {
		key = ck.String()
	}
	result = redis.GetInstance(ck.ConfigName).Exists(key)
	return result
}
func ExistsMulti(ck []CacheKey) ([]CacheKey, []CacheKey) {
	ResultTrue := []CacheKey{}
	ResultFalse := []CacheKey{}
	keys := map[string][]string{}
	for _, v := range ck {
		if len(keys[v.ConfigName]) == 0 {
			keys[v.ConfigName] = []string{}
		}
		if v.CType == CacheTypeRelation {
			keys[v.ConfigName] = append(keys[v.ConfigName], v.GetCacheKey())
		} else {
			keys[v.ConfigName] = append(keys[v.ConfigName], v.String())
		}
	}
	for k, v := range keys {
		vv := tools.RemoveDuplicateElement(v)
		tmp := redis.GetInstance(k).ExistsMulti(vv)
		for _, val := range ck {
			key := ""
			if val.CType == CacheTypeRelation {
				key = val.GetCacheKey()
			} else {
				key = val.String()
			}
			if _, ok := tmp[key]; ok {
				if tmp[key] {
					ResultTrue = append(ResultTrue, val)
				} else {
					ResultFalse = append(ResultFalse, val)
				}
			}
		}
	}
	return RemoveDuplicateCacheKey(ResultTrue), RemoveDuplicateCacheKey(ResultFalse)
}
func Incr(ck CacheKey, val int) (int, error) {
	var result int
	var err error
	val = tools.Abs(val)
	if ck.CType == CacheTypeRelation {
		if redis.GetInstance(ck.ConfigName).HExists(ck.GetCacheKey(), ck.Params[1]) {
			result = redis.GetInstance(ck.ConfigName).HIncr(ck.GetCacheKey(), ck.Params[1], val)
		} else {
			err = errors.New("not exists")
		}
	} else {
		if redis.GetInstance(ck.ConfigName).Exists(ck.String()) {
			result = redis.GetInstance(ck.ConfigName).IncrBy(ck.String(), val)
		} else {
			err = errors.New("not exists")
		}
	}
	return result, err
}
func Decr(ck CacheKey, val int) (int, error) {
	var result int
	var err error
	if ck.CType == CacheTypeRelation {
		res := redis.GetInstance(ck.ConfigName).HMGet(ck.GetCacheKey(), ck.Params[1])
		v, _ := strconv.Atoi(res[ck.Params[1]])
		if v >= val {
			result = redis.GetInstance(ck.ConfigName).HDecr(ck.GetCacheKey(), ck.Params[1], val)
		} else {
			err = errors.New("not enough")
		}
	} else {
		res := redis.GetInstance(ck.ConfigName).Get(ck.String())
		v, _ := strconv.Atoi(res)
		if v >= val {
			result = redis.GetInstance(ck.ConfigName).DecrBy(ck.String(), val)
		} else {
			err = errors.New("not enough")
		}
	}
	return result, err
}
func FixDate(data string) int {
	return int(tools.StrToTime(data) - time.Now().Unix())
}
func SetRealCacheDataDefault(ck []CacheKey, rcd []RealCacheData) []RealCacheData {
	i := 1
	for _, val := range ck {
		for _, v := range rcd {
			if val.String() == v.CacheKey.String() {
				i = 2
			}
		}
		if i == 1 {
			rcd = append(rcd, RealCacheData{val.DefaultVal, val})
		}
	}
	return rcd
}
