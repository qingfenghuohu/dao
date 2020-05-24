package dao

import (
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
	for key, val := range dataCacheKey {
		switch key {
		case UserProductModelDataCacheKeyIntegral:
			uid := []string{}
			_, ck := ExistsMulti(val)
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
				uid = tools.RemoveDuplicateElement(uid)
				wh := strings.Join(uid, ",")
				res := Model(u).Where("uid in(" + wh + ") and type = 1").Select()
				for _, v := range res {
					rcd := RealCacheData{}
					rcd.CacheKey = CreateCacheKey(u, UserProductModelDataCacheKeyIntegral, v["Uid"].(string), v["Pid"].(string))
					rcd.Result = v["Val"].(string)
					result = append(result, rcd)
				}
			}
			result = SetRealCacheDataDefault(val, result)
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
				res := Model(u).Where("uid in(" + wh + ") and type = 2").Select()
				for _, v := range res {
					rcd := RealCacheData{}
					rcd.CacheKey = CreateCacheKey(u, UserProductModelDataCacheKeyIntegral, v["Uid"].(string), v["Pid"].(string))
					rcd.Result = v["Val"].(string)
					result = append(result, rcd)
				}
			}
			result = SetRealCacheDataDefault(val, result)
			break
		}
	}
	return result
}

func (u *UserProduct) DbToCache(md ModelData, ck []CacheKey) RealData {
	var result RealData
	for _, val := range md.Data {
		for _, v := range ck {
			switch v.Key {
			case UserProductModelDataCacheKeyIntegral, UserProductModelDataCacheKeyNum:
				if md.Operation == "del" {
					v.Params = append(v.Params, val.BeData["Uid"].(string))
					v.Params = append(v.Params, val.BeData["Pid"].(string))
					result.DelAppend(v)
				} else {
					if val.BeData["Pid"].(string) == val.AfterData["Pid"].(string) {
						v.Params = []string{}
						v.Params = append(v.Params, val.BeData["Uid"].(string))
						v.Params = append(v.Params, val.BeData["Pid"].(string))
						result.DelAppend(v)
					}
					v.Params = []string{}
					v.Params = append(v.Params, val.AfterData["Uid"].(string))
					v.Params = append(v.Params, val.AfterData["Pid"].(string))
					result.Append(RealCacheData{val.AfterData["Values"].(string), v})
				}
				break
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
		DefaultVal: nil,
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
		DefaultVal: nil,
	}
	return result
}
