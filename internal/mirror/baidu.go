package mirror

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// 百度地图服务端 AK（取自 mirror_frontend AppConfig.baiduMapServerKey，
// 与 app 端地点选择走同一把 key、同一接口、同一坐标系 bd09ll）。
const baiduServerAK = "Ky1ZyM6hPNf5WRmVfUzMM0MLcOBSUKHi"

type Place struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Province string `json:"province"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Location struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
}

type baiduPlaceResponse struct {
	Status  int     `json:"status"`
	Message string  `json:"message"`
	Results []Place `json:"results"`
}

// SearchPlaces 百度国内地点检索（与 app 端 BdMapUtils.searchDomesticPlaces 同参）。
// 返回坐标为百度 bd09ll 坐标系——与 app 选点入库的坐标系一致。
func SearchPlaces(query, region string) ([]Place, error) {
	q := url.Values{
		"query":     {query},
		"region":    {region},
		"page_size": {"30"},
		"page_num":  {"0"},
		"scope":     {"1"},
		"output":    {"json"},
		"ak":        {baiduServerAK},
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).
		Get("https://api.map.baidu.com/place/v3/region?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out baiduPlaceResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("百度地点检索响应非 JSON: %.200s", raw)
	}
	if out.Status != 0 {
		return nil, fmt.Errorf("百度地点检索失败 status=%d message=%s", out.Status, out.Message)
	}
	return out.Results, nil
}
