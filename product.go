package dao

type Product struct {
	Id      int     `gorm:"type:int(11) NOT NULL AUTO_INCREMENT;PRIMARY_KEY;column:id;"` //自增id
	Name    string  `gorm:"type:varchar(255);column:name;"`                              //产品名称
	Day     int     `gorm:"type:int(10) ;column:day;"`                                   //有效天数
	Price   float64 `gorm:"type:decimal(10, 2);column:price;"`                           //价格
	Content string  `gorm:"type:varchar(255);column:content;"`                           //产品值
	Type    int     `gorm:"type:tinyint(1);column:type;"`                                //类型1=积分,2=次数,3=多品,4=自由
	State   int     `gorm:"type:tinyint(1);DEFAULT:0;column:state;"`                     //状态1=开启
	Score   int     `gorm:"type:int(11);DEFAULT:0;column:score;"`                        //排序积分
	Address string  `gorm:"type:varchar(255);column:address;"`                           //地址Url
	Term    string  `gorm:"type:varchar(255);DEFAULT:0;column:term;"`                    //约束 例如:自由的次数
	Details string  `gorm:"type:text;DEFAULT:0;column:details;"`                         //详情
}

var AddressPidKey = "addressPid:%s"

const (
	ProductModelDataCacheKeyState = "s"
)

func (u *Product) MicroName() string {
	return "p"
}
func (u *Product) TableName() string {
	return "product"
}

func (u *Product) DbName() string {
	return "third"
}

func (u *Product) GetRealData(dataCacheKey map[string][]CacheKey) []RealCacheData {
	var result []RealCacheData
	return result
}

func (u *Product) DbToCache(md ModelData, ck []CacheKey) RealData {
	var result RealData
	return result
}

func (u *Product) GetDataCacheKey() map[string]CacheKey {
	result := make(map[string]CacheKey)
	result[CacheTypeIds] = CacheKey{
		Key:        CacheTypeIds,
		CType:      CacheTypeIds,
		Model:      u,
		LifeTime:   3600 * 24 * 30,
		ResetTime:  0,
		ResetCount: 0,
		Version:    1,
		RelField:   []string{"Id"},
		ResetType:  1,
		ConfigName: u.DbName(),
	}
	result[ProductModelDataCacheKeyState] = CacheKey{
		Key:        ProductModelDataCacheKeyState,
		CType:      CacheTypeField,
		LifeTime:   3600 * 24 * 30,
		Model:      u,
		ResetTime:  0,
		ResetCount: 0,
		Version:    1,
		ResetType:  1,
		RelField:   []string{"State"},
		ConfigName: u.DbName(),
	}
	return result
}
