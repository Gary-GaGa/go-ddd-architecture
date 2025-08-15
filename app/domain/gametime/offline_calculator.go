package gametime

import (
	"time"

	"go-ddd-architecture/app/domain/player"
)

const (
	maxOffline = 8 * time.Hour
)

// Clock 介面可替換實作以利測試。
type Clock interface {
	Now() time.Time
}

// OfflineResult 計算結果。
type OfflineResult struct {
	GainedKnowledge int64
	GainedResearch  int64
	ClampedTo8h     bool
	AnomalyDetected bool
	Message         string
}

// OfflineCalculator 根據關閉時與現在的時間，計算離線收益。
// MVP：以固定產率做簡化，後續再接入語言/等級影響。
type OfflineCalculator struct {
	// base per-minute gains as MVP constants
	knowledgePerMinute int64
	researchPerMinute  int64
}

func NewOfflineCalculator() *OfflineCalculator {
	return &OfflineCalculator{
		knowledgePerMinute: 10, // MVP 常數，可調整
		researchPerMinute:  2,
	}
}

// Compute 計算並應用到玩家（不直接持久化）。
func (c *OfflineCalculator) Compute(p *player.Player, ts Timestamps, now time.Time) OfflineResult {
	dtWall := now.Sub(ts.WallClockAtClose)
	if dtWall <= 0 {
		// 時間倒退或無進展 → 判為異常，0 收益
		p.LastSeen = now
		return OfflineResult{AnomalyDetected: true, Message: "time anomaly detected; no offline gains"}
	}

	dt := dtWall
	clamped := false
	if dt > maxOffline {
		dt = maxOffline
		clamped = true
	}

	// 簡化：忽略單調代理值的精細比對，先做 MVP：若代理值為 0，跳過檢查。
	// 後續可加入期望差值門檻，嚴重不符改為 0 或數分鐘上限。

	minutes := int64(dt / time.Minute)
	if minutes <= 0 {
		p.LastSeen = now
		return OfflineResult{}
	}

	// 若玩家有等級，依玩家產率計算；否則用常數
	kpm := c.knowledgePerMinute
	rpm := c.researchPerMinute
	if p != nil {
		if v := p.KnowledgePerMinute(); v > 0 {
			kpm = v
		}
		if v := p.ResearchPerMinute(); v > 0 {
			rpm = v
		}
	}
	gk := minutes * kpm
	gr := minutes * rpm

	p.ApplyOfflineGains(gk, gr, now)

	return OfflineResult{
		GainedKnowledge: gk,
		GainedResearch:  gr,
		ClampedTo8h:     clamped,
	}
}
