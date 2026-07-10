package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/BIXING-CODE/winik-cli/internal/config"
	"github.com/BIXING-CODE/winik-cli/internal/mirror"
)

type multiFlag []string

func (m *multiFlag) String() string     { return fmt.Sprint(*m) }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

// Action 发布一条行动：本地图片自动走 COS 上传 + fragment 注册。
func Action(args []string) error {
	fs := flag.NewFlagSet("action", flag.ExitOnError)
	title := fs.String("title", "", "行动标题（必填）")
	content := fs.String("content", "", "行动内容（必填）")
	cover := fs.String("cover", "", "封面图本地路径（与 --cover-url 二选一，必填其一）")
	coverURL := fs.String("cover-url", "", "封面图已有 URL")
	var images multiFlag
	fs.Var(&images, "image", "附加图片本地路径（可重复）")
	timeDesc := fs.String("time-desc", "", "时间文字描述（与 --start-at 二选一，必填其一）")
	startAt := fs.String("start-at", "", `具体开始时间 "2026-07-15 19:00"`)
	location := fs.String("location", "", "线下地点名称（手动模式，需配 --lat/--lng）")
	lat := fs.Float64("lat", 0, "纬度（bd09ll 百度坐标系）")
	lng := fs.Float64("lng", 0, "经度（bd09ll 百度坐标系）")
	place := fs.String("place", "", "地点关键词，自动经百度检索取第一个 POI（推荐，先用 place 命令预览）")
	placeCity := fs.String("place-city", "北京", "配合 --place 的检索城市")
	placeIndex := fs.Int("place-index", 0, "取百度检索第 N 个结果（place 命令输出的序号）")
	online := fs.Bool("online", false, "线上碰面（默认线下，线下需 --location/--lat/--lng）")
	price := fs.String("price", "", "价格（必填，字符串）")
	selfOnly := fs.Bool("self", false, "仅自己可见（默认公开）")
	asJSON := fs.Bool("json", false, "输出原始 JSON 响应")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// 与 app 端 validateInfos 对齐的最小校验
	if *title == "" || *content == "" || *price == "" {
		return fmt.Errorf("--title / --content / --price 必填")
	}
	if *cover == "" && *coverURL == "" {
		return fmt.Errorf("--cover 或 --cover-url 必填其一")
	}
	if *timeDesc == "" && *startAt == "" {
		return fmt.Errorf("--time-desc 或 --start-at 必填其一")
	}
	if *startAt != "" {
		if _, err := time.Parse("2006-01-02 15:04", *startAt); err != nil {
			return fmt.Errorf(`--start-at 格式须为 "2026-07-15 19:00": %w`, err)
		}
	}
	if !*online && *location == "" && *place == "" {
		return fmt.Errorf("线下行动需 --place（自动解析）或 --location + --lat/--lng（或改用 --online）")
	}
	if !*online && *place == "" && (*lat == 0 || *lng == 0) {
		return fmt.Errorf("手动指定 --location 时必须同时给 --lat/--lng（否则 app 端点开地点会报错）；推荐改用 --place 自动解析")
	}

	// --place: 百度地点检索解析出真实 POI（名称 + bd09ll 经纬度），与 app 选点同源
	if *place != "" && !*online {
		places, err := mirror.SearchPlaces(*place, *placeCity)
		if err != nil {
			return fmt.Errorf("百度地点检索失败: %w", err)
		}
		if len(places) == 0 {
			return fmt.Errorf("「%s」在 %s 无检索结果，换 --place-city 或先用 place 命令预览", *place, *placeCity)
		}
		if *placeIndex >= len(places) {
			return fmt.Errorf("--place-index %d 超出结果数 %d", *placeIndex, len(places))
		}
		p := places[*placeIndex]
		*location = p.Name
		*lat, *lng = p.Location.Lat, p.Location.Lng
		fmt.Printf("地点解析: %s（%s%s%s · %s）lat=%.6f lng=%.6f\n",
			p.Name, p.Province, p.City, p.Area, p.Address, *lat, *lng)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Bixing.Token == "" {
		return fmt.Errorf("未登录 bixing，先执行 winik-cli login bixing --phone <手机号> 或 --token <token>")
	}
	client := mirror.New(cfg.Bixing.BaseURL, cfg.Bixing.Token)

	// 1. 封面
	coverImage := *coverURL
	if *cover != "" {
		fmt.Printf("上传封面 %s ...\n", *cover)
		frag, err := client.UploadPicture(*cover)
		if err != nil {
			return fmt.Errorf("封面上传失败: %w", err)
		}
		coverImage = frag.SourceURL
	}

	// 2. 附加图片
	var fragmentIDs []int
	for _, img := range images {
		fmt.Printf("上传图片 %s ...\n", img)
		frag, err := client.UploadPicture(img)
		if err != nil {
			return fmt.Errorf("图片 %s 上传失败: %w", img, err)
		}
		if frag.ID != nil {
			fragmentIDs = append(fragmentIDs, *frag.ID)
		}
	}

	// 3. 组装发布
	mode := 0 // OFFLINE
	if *online {
		mode = 1 // ONLINE
	}
	visible := 1 // public
	if *selfOnly {
		visible = 0 // self
	}
	req := &mirror.ActionRequest{
		Title:         *title,
		Content:       *content,
		CoverImage:    coverImage,
		Location:      *location,
		Price:         *price,
		TimeRangeDesc: *timeDesc,
		FragmentIDs:   fragmentIDs,
		Mode:          &mode,
		VisibleStatus: &visible,
	}
	if *startAt != "" {
		req.StartAt = *startAt + ":00" // 线格式 yyyy-MM-dd HH:mm:ss
		req.TimeRangeDesc = ""
	}
	if *lat != 0 || *lng != 0 {
		req.Lat, req.Lng = lat, lng
	}

	resp, err := client.PublishAction(req)
	if err != nil {
		return fmt.Errorf("发布失败: %w", err)
	}
	if *asJSON {
		out, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	id := 0
	if resp.ID != nil {
		id = *resp.ID
	}
	fmt.Printf("发布成功 id=%d 审核状态=%s 标题=%s\n", id, resp.ActionStatus, resp.Title)
	return nil
}
