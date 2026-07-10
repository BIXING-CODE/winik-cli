package cmd

import (
	"flag"
	"fmt"

	"github.com/BIXING-CODE/winik-cli/internal/mirror"
)

// Place 百度地点检索预览：列出候选 POI（名称/地址/经纬度），供发行动前确认。
func Place(args []string) error {
	fs := flag.NewFlagSet("place", flag.ExitOnError)
	query := fs.String("query", "", "地点关键词（必填）")
	city := fs.String("city", "北京", "检索城市")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *query == "" {
		return fmt.Errorf("--query 必填")
	}
	places, err := mirror.SearchPlaces(*query, *city)
	if err != nil {
		return err
	}
	if len(places) == 0 {
		fmt.Printf("「%s」在 %s 无结果，试试换城市（--city）或换关键词\n", *query, *city)
		return nil
	}
	for i, p := range places {
		fmt.Printf("[%d] %s\n    %s%s%s · %s\n    lat=%.6f lng=%.6f\n",
			i, p.Name, p.Province, p.City, p.Area, p.Address, p.Location.Lat, p.Location.Lng)
	}
	return nil
}
