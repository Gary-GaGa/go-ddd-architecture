# Golang DDD × Clean Architecture × 工程師訓練 AI 寵物放置型遊戲

> **專案定位**  
> 以 Golang 實踐 DDD 與 Clean Architecture，打造一款工程師訓練 AI 寵物的放置型遊戲，結合 AI 輔助，專注於架構設計、遊戲核心邏輯與現代工程流程。



## 專案簡介

本專案以「工程師訓練 AI 寵物」為主題，玩家扮演軟體工程師，培養專屬 AI 寵物，透過學習程式語言、解題、升級技能樹，體驗放置型成長循環。  
融合 DDD、Clean Architecture 與 Golang 實踐，並善用 AI 工具提升設計與開發效率。

> [!TIP]
> 本專案同時作為 DDD/Clean Architecture 教學範例與遊戲設計實驗場，適合工程師、遊戲設計師、AI 協作愛好者參考。


---



## 目標與特色

- 以 Golang 為主，實踐 DDD 聚合根、實體、值物件、服務、Repository 分層
- 採用 Clean Architecture、Hexagonal 架構，嚴格區分 Domain、Application、Adapter、Infrastructure
- 前端以 Ebitengine（Ebiten）作為 Presenter/UI，透過 Usecase Port 介接邏輯；不串接後端（完全離線，本地存檔）
- 遊戲主題：工程師培養 AI 寵物，放置型成長循環（語言學習 → 解題 → 收益 → 升級 → 新挑戰）
- 系統模組：技能樹、資源系統（知識點、研發點、算力）、升級模組、語言學習、任務系統
- 運用 AI（如 ChatGPT）輔助企劃、架構設計、測試案例與程式生成
- 著重架構設計、邏輯正確性，Domain Core 可獨立單元測試
- 以「遊戲設計文檔」驅動程式架構，設計流程與程式碼同步演進

---


## 專案目錄結構


```
/domain      # DDD：純領域模型與介面
/adapter     # 介面轉接層、repository 實作
/infra       # Infrastructure，底層技術細節
/usecase     # 應用層：協調流程
/app         # Application 層
/cmd         # 入口與啟動
/docs        # 遊戲設計、架構說明
...
```

---



## 遊戲設計概要

- **主題**：工程師訓練 AI 寵物，AI 透過學習程式語言解鎖技能並自動解題
- **核心循環**：語言學習 → 解題 → 收益 → 升級 → 新挑戰 → 轉生循環
- **系統模組**：
  - 技能樹（以語言為主軸分支，解鎖專屬技能）
  - 資源系統（知識點、研發點、算力）
  - 升級模組（AI 硬體與軟體升級）
  - 語言學習系統（循序解鎖，含熟練度與練習任務）
  - 任務系統（自動解題、指定任務、綜合挑戰任務）
- **UI 構想**：主介面、語言學習、技能樹升級、任務日誌、離線收益
- **程式語言與技能對應表**：

| 語言 | 技能 |
|------|------|
| Python | 資料處理、機器學習、自動化腳本 |
| C++ | 高效算法、記憶體優化、系統控制 |
| Java | 並行處理、後端架構、大型專案模擬 |
| JavaScript | UI 模擬、網頁互動、資料擷取 |
| Go | 高併發服務、API 模擬、伺服器任務 |

- **放置與解題邏輯**：AI 學會語言後即可自動解題，產生知識點、研發點與特殊事件，離線收益可累積 8 小時

> [!NOTE]
> 詳細遊戲設計請參考 `docs/game-design.md`。

---


## 開發流程

1. 依據遊戲設計文檔規劃 Domain Model
2. 建立分層架構，撰寫核心邏輯與測試
3. 逐步實作遊戲模組（AI 寵物、語言學習、技能樹、任務系統等）
4. 善用 AI 工具協作設計、生成程式碼與測試
5. 持續優化架構、補充文檔與測試案例

---



## 適合對象

- 想理解 DDD 在遊戲開發中的實踐
- 有志於用現代工程方法開發創意產品
- 希望學習 AI 協作於軟體設計流程
- 對放置型遊戲、AI 寵物、程式語言學習主題有興趣

---



## 施工中

- 持續完善聚合根、核心流程設計
- 陸續更新遊戲循環、技能樹、語言學習、測試案例與技術文檔

---



## 參考資源

- DDD、Clean Architecture 經典文獻
- Golang 軟體設計範例
- AI 輔助軟體工程實踐

---

如需更詳細的遊戲設計與架構說明，請參考：

- `docs/game-design.md`
- `docs/architecture.md`
- `docs/ebiten-integration.md`（Ebiten 作為外部 App，透過後端 Adapter 串接的設計）
- `docs/http-adapter.md`（HTTP 介面設計：ViewModel 與離線結算端點）
- `docs/art-style.md`（美術風格與 UI 指南）

---


## MVP 里程碑與開發任務

### MVP 目標（2 週內可玩）
- Go/Python 兩條語言
- 練習任務與解題任務
- 知識點與研發點資源系統
- 升級語言等級
- 離線收益上限 8 小時
- bbolt 儲存
- CLI 與 Ebitengine 簡易 UI
 - 純本地時間校驗（wall-clock + 單調時間代理值），偵測時間異常時採保守結算並提示

### 開發任務
- Sprint 0：建立 Domain 模型與用例、離線收益計算、bbolt 儲存庫實作、CLI 介面、單元測試。
- Sprint 1：Ebitengine 桌面版主迴圈（Update/Draw）、UI Presenter、按鍵互動、存讀檔整合。
- 驗收標準：啟動可查看資源與任務狀態、開始任務並完成獲獎、升級語言後產出變化、離線收益正確計算並受 8 小時限制、存檔可於關閉後重啟讀取、時間逆轉/異常跳動時離線收益採保守結算並提示。

## 專案目錄規劃（MVP 階段）

```
/domain
  /player
  /task
  /language
  /resource
  clock.go
  rng.go
  events.go
/usecase
  game_loop.go
/adapter
  presenter/
  repo/
/infra
  /storage/bbolt_store.go
/cmd
  /cli/main.go
  /desktop/main.go
/configs/*.json
/docs
```

- `/domain`：純領域邏輯，包含玩家、任務、語言、資源等核心模型與事件。
- `/usecase`：應用服務層，負責遊戲主流程與業務協調。
- `/adapter`：轉接層，包含 Presenter 與 Repository 實作，負責介面與資料存取。
- `/infra`：基礎設施層，實作底層技術細節，如 bbolt 資料庫存取。
- `/cmd`：應用啟動程式碼，包含 CLI 與桌面版入口。
- `/configs`：配置檔案，存放 JSON 格式的設定。
- `/docs`：遊戲設計與架構說明文件。

---

## 架構草圖

請參考 `docs/architecture.md`，內含：
- 層次/元件原理圖（Presenter × Usecase × Domain × Infra）
- 啟動與離線收益序列圖
- Update/Draw 主迴圈與 Advance(Δt) 時間驅動
- 介面約定、領域模型速覽、離線時間校驗策略

---

## 快速啟動：Server + Ebiten Client

以下步驟會在本機啟動 HTTP 伺服器（預設 `127.0.0.1:8080`）與桌面版 Ebiten Client。

1) 啟動後端伺服器（Cobra CLI）

```bash
# 在專案根目錄
go run ./cmd/cli server

# 參數：
# --mem   使用記憶體儲存庫（預設 true，不落地存檔）
# --db    指定 bbolt 檔案路徑（需搭配 --mem=false 才會使用）
# 範例（使用 bbolt 落地存檔）
# go run ./cmd/cli server --mem=false --db=game.db
```

2) 啟動 Ebiten Client（桌面視窗）

```bash
# 方法 A：快速建構與執行
go run ./client/cmd/ebiten-client

# 方法 B：建構二進位後執行
go build -o bin/ebiten-client ./client/cmd/ebiten-client
./bin/ebiten-client
```

啟動後，視窗左側會顯示資源與任務資訊，右側為語言主題的 AI 視覺區。

### 操作鍵（Controls）

- P：開始練習任務（Practice）
- T：開始 Targeted 任務（短時、略高獎勵）
- U：升級 Knowledge（消耗 Research；不足時會提示）
- C：結算離線收益（Claim Offline）
- 1/2/3：切換語言（Go / Python / JavaScript）

任務行為：

- 當沒有進行中任務時，會自動開始練習（Practice）。
- 任務倒數至 0 後，會自動嘗試結算（Try Finish），不用手動按鍵。

提示：

- 左側 Status 卡會顯示 Est. Success（估計成功率）、Rates（每分鐘知識/研發產率）。
- 中央任務卡會顯示語言、Base 奬勵與總時長，右側圓環以 mm:ss 倒數。
- 中央任務卡背景有淡藍藍圖格線與任務型別標籤（Practice/Targeted）。
- 右側 Languages 面板每列提供升級進度條（Research / Next Cost），並在可升級時顯示 READY 徽章；當前選中的語言以強調色顯示。

如遇連線狀態，左側會短暫顯示 Networking...，錯誤則以 Error 行顯示；升級不足會有 toast 提示。

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Disclaimer

Some game mechanics or ideas are inspired by existing idle/incremental games.  
This project is non-commercial and created for educational purposes only.  