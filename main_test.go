package dao

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestModel3(t *testing.T) {
	UserProduct := UserProduct{}
	Model(&UserProduct).Where("id = 1").Del()

	//UserProduct.Type = 1
	//UserProduct.Uid = 1
	//UserProduct.Pid = tools.MtRand(1,10000)
	//UserProduct.Val = tools.MtRand(100,10000)
	//Model(&UserProduct).Where("id = 1").Save(&UserProduct)

	//Model(&UserProduct).Add(&UserProduct)
	//Product := Product{}
	//result := GetCache().
	//	Add(&Product, CacheTypeIds, "1").
	//	Add(&UserProduct, UserProductModelDataCacheKeyIntegral, "1", "2").
	//	Add(&UserProduct, UserProductModelDataCacheKeyIntegral, "1", "3").
	//	GetData()
	//fmt.Println(result.Read(&Product, CacheTypeIds, "1").Map())
	//res := result.Read(&UserProduct, UserProductModelDataCacheKeyIntegral, "1", "3").Int()
	//fmt.Println(res)
}
func TestModel2(t *testing.T) {
	//st := "tttttttttttttttttttttttttttfdfdsfdfdvewereqewwqewq"
	//var res interface{}
	//v := int(crc32.ChecksumIEEE([]byte(st)))
	//if v >= 0 {
	//	res =  v
	//}
	//if -v >= 0 {
	//	res = -v
	//}
	//fmt.Println(res)
	// v == MinInt
	st := "123kikI"
	var seed uint64 = 0
	var magicNumber uint64 = 0x9e3779b9
	for _, s := range st {
		seed ^= uint64(s) + magicNumber + (seed << 6) + (seed >> 2)
	}

	fmt.Println(st, seed, uint16(seed), strconv.Itoa(int(uint16(seed))))
}

func TestModel(t *testing.T) {
	//str := []string{"1111"}
	//f := func() {
	//	var i int
	//	for i = 0; i < 10; i++ {
	//		str = append(str,strconv.Itoa(i)+"ww")
	//	}
	//}
	//f()
	//fmt.Println(str)
	Product := Product{}
	//flag := Model(&Product).Where("id = 2").Del()
	//fmt.Println(flag)

	//Product.Name = "test"
	//Product.Content = strconv.Itoa(tools.MtRand(1000,9999))
	//flag := Model(&Product).Where("id = 1").Add(&Product)
	//fmt.Println(flag)

	//Product.Content = strconv.Itoa(tools.MtRand(1000,9999))
	//Product.Price = 0
	//flag := Model(&Product).Where("id = 1").Save(&Product)
	//fmt.Println(flag)

	//var config []DataCacheKey
	//config = append(config,CreateCacheKey(&Product, ProductModelDataCacheKeyState, "1"))
	//res := tools.Interface2MapSliceStr(GetData(&config))
	//fmt.Println(res)

	tt := time.Now()
	result := GetCache().
		Add(&Product, CacheTypeIds, "10").
		Add(&Product, CacheTypeIds, "1").
		Add(&Product, ProductModelDataCacheKeyState, "1").
		Add(&Product, ProductModelDataCacheKeyState, "2").
		GetData()
	fmt.Println(result.Read(&Product, ProductModelDataCacheKeyState, "2").SliceMap())
	elapsed := time.Since(tt)
	fmt.Println("main_test:", elapsed)

	//result := redis.GetInstance("third").MGet("h","1","ooo")
	//fmt.Println(result)
}
