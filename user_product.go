package dao

import (
	"fmt"
	"github.com/qingfenghuohu/tools"
	"strings"
	"time"
)

type UserProduct struct {
	Id        int `gorm:"type:int(11) NOT NULL AUTO_INCREMENT;PRIMARY_KEY;column:id;"` //自增id
	Uid       int `gorm:"type:int(11);column:uid;"`                                    //用户id
	StartTime int `gorm:"type:int(11);column:start_time;"`                             //开始时间
	EndTime   int `gorm:"type:int(11);column:end_time;"`                               //结束时间
	Val       int `gorm:"type:int(11);column:val;"`                                    //产品具体值
	Pid       int `gorm:"type:int(11);column:pid;"`                                    //产品id
	Type      int `gorm:"type:tinyint(1);DEFAULT:0;column:type;"`                      //产品类型 1=积分,2=次数
}

const (
	UserProductModelDataCacheKeyUid      = "uid"
	UserProductModelDataCacheKeyIntegral = "integral"
	UserProductModelDataCacheKeyNum      = "num"
)

func (u *UserProduct) MicroName() string {
	return "p"
}
func (u *UserProduct) TableName() string {
	return "user_product"
}

func (u *UserProduct) DbName() string {
	return "third"
}

func (u *UserProduct) GetRealData(dataCacheKey map[string][]CacheKey) []RealCacheData {
	var result []RealCacheData
	res := map[string]RealCacheData{}
	for key, val := range dataCacheKey {
		switch key {
		case UserProductModelDataCacheKeyIntegral:
			uid := []string{}
			for _, v := range val {
				res[v.String()] = RealCacheData{v.DefaultVal, v}
			}
			_, ck := ExistsMulti(val)
			for _, v := range ck {
				if v.Params[0] != "" {
					uid = append(uid, v.Params[0])
				}
			}
			if len(uid) > 0 {
				uid = tools.RemoveDuplicateElement(uid)
				wh := strings.Join(uid, ",")
				dbdata := Model(u).Where("uid in(" + wh + ") and type = 1").Select()
				for _, v := range dbdata {
					k := CreateCacheKey(u, UserProductModelDataCacheKeyIntegral, v["Uid"].(string), v["Pid"].(string))
					res[k.String()] = RealCacheData{v["Val"].(string), k}
				}
			}
			break
		case UserProductModelDataCacheKeyNum:
			uid := []string{}
			t, ck := ExistsMulti(val)
			for _, v := range t {
				rcd := RealCacheData{v.DefaultVal, v}
				result = append(result, rcd)
			}
			for _, v := range ck {
				if !Exists(v) {
					if v.Params[0] != "" {
						uid = append(uid, v.Params[0])
					}
				} else {
					rcd := RealCacheData{v.DefaultVal, v}
					result = append(result, rcd)
				}
			}
			if len(uid) > 0 {
				wh := strings.Join(uid, ",")
				dbdata := Model(u).Where("uid in(" + wh + ") and type = 2").Select()
				for _, v := range dbdata {
					k := CreateCacheKey(u, UserProductModelDataCacheKeyIntegral, v["Uid"].(string), v["Pid"].(string))
					res[k.String()] = RealCacheData{v["Val"].(string), k}
				}
			}
			break
		}
	}
	for _, v := range res {
		result = append(result, v)
	}
	return result
}

func (u *UserProduct) DbToCache(md *ModelData, ck []CacheKey) RealData {
	var result RealData
	for _, val := range md.Data {
		fmt.Println(val)
		if md.Operation == "del" {
			if val.BeData["Type"].(string) == "1" {
				result.DelAppend(CreateCacheKey(md.Model, UserProductModelDataCacheKeyIntegral, val.BeData["Uid"].(string), val.BeData["Pid"].(string)))
			}
			if val.BeData["Type"].(string) == "2" {
				result.DelAppend(CreateCacheKey(md.Model, UserProductModelDataCacheKeyNum, val.BeData["Uid"].(string), val.BeData["Pid"].(string)))
			}
		} else {
			if val.BeData["Type"].(string) == "1" {
				result.DelAppend(CreateCacheKey(md.Model, UserProductModelDataCacheKeyIntegral, val.BeData["Uid"].(string), val.BeData["Pid"].(string)))
			}
			if val.BeData["Type"].(string) == "2" {
				result.DelAppend(CreateCacheKey(md.Model, UserProductModelDataCacheKeyNum, val.BeData["Uid"].(string), val.BeData["Pid"].(string)))
			}
			if val.AfterData["Type"].(string) == "1" {
				key := CreateCacheKey(md.Model, UserProductModelDataCacheKeyIntegral, val.AfterData["Uid"].(string), val.AfterData["Pid"].(string))
				if Exists(key) {
					result.Append(RealCacheData{val.AfterData["Val"].(string), key})
				} else {
					result.SaveAppend(key)
				}
			}
			if val.AfterData["Type"].(string) == "2" {
				key := CreateCacheKey(md.Model, UserProductModelDataCacheKeyNum, val.AfterData["Uid"].(string), val.AfterData["Pid"].(string))
				if Exists(key) {
					result.Append(RealCacheData{val.AfterData["Val"].(string), key})
				} else {
					result.SaveAppend(key)
				}
			}
		}
	}
	return result
}

func (u *UserProduct) GetDataCacheKey() map[string]CacheKey {
	result := make(map[string]CacheKey)
	result[UserProductModelDataCacheKeyIntegral] = CacheKey{
		Key:        UserProductModelDataCacheKeyIntegral,
		CType:      CacheTypeRelation,
		Model:      u,
		LifeTime:   3600 * 24 * 30,
		ResetTime:  0,
		ResetCount: 0,
		Version:    1,
		RelField:   []string{"Uid", "Pid"},
		ConfigName: u.DbName(),
		DefaultVal: "",
	}
	result[UserProductModelDataCacheKeyNum] = CacheKey{
		Key:        UserProductModelDataCacheKeyNum,
		CType:      CacheTypeRelation,
		Model:      u,
		LifeTime:   FixDate(tools.Date(time.Now().Unix(), "2006-01-02") + " 23:59:59"),
		ResetTime:  0,
		ResetCount: 0,
		Version:    1,
		RelField:   []string{"Uid", "Pid"},
		ConfigName: u.DbName(),
		DefaultVal: "",
	}
	return result
}
