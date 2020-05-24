package dao

type Real struct {
	dck []CacheKey
}

func (real *Real) SetCacheData(rcd []RealCacheData) {
}

func (real *Real) GetCacheData(res *Result) {

}

func (real *Real) GetRealData() []RealCacheData {
	var result []RealCacheData
	return result
}

func (real *Real) SetDataCacheKey(dck []CacheKey) Cache {
	real.dck = RemoveDuplicateCacheKey(dck)
	return real
}
func (real *Real) DelCacheData(dck []CacheKey) {

}
func (real *Real) DbToCache(md ModelData) []RealCacheData {
	var result []RealCacheData
	return result
}
func (real *Real) GetCacheKey(key *CacheKey) string {
	return key.String()
}
