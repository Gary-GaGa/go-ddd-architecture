# HTTP Adapter 規劃（Ebiten/前端 對接）

本文件規劃以標準 `net/http` 實作的 HTTP Adapter，供本地 Ebiten（或其他前端）呼叫 Usecase。維持 Clean Architecture：Adapter 只依賴 Port（Usecase 介面與 DTO），不滲入領域細節。

## 目標
- 提供最小 API（已實作，採用版本化前綴 /api/v1）：
  - GET /api/v1/game/viewmodel 取得展示資料
  - POST /api/v1/game/claim-offline 結算離線收益
  - POST /api/v1/game/start-practice 開始練習任務
  - POST /api/v1/game/try-finish 嘗試完成當前任務
- 可選擴充：模擬時間與回推關閉時間（便於測試/開發）
- 簡單、無認證（本地開發用），未來可加上 Token 或 IPC

## 端點設計

> 相容策略：暫時保留舊版 /api/* 路徑作為過渡（與 /api/v1/* 同義）。

## Middleware
- Request ID：每個請求在 Header X-Request-Id 傳遞，若未提供則由伺服器產生。
- Recovery：攔截 panic，回應 500 並記錄日誌。
- Access Log：輸出 method、path、status、duration、request id（標準 log）。

## 端點設計

### GET /api/v1/game/viewmodel
- 說明：回傳目前的簡化 ViewModel。
- 回傳 200 JSON：
```
{
  "knowledge": 12345,
  "research": 678,
  "notices": ["...optional..."]
}
```
- 型別對應：`app/usecase/dto/game.ViewModelDto`

### POST /api/v1/game/claim-offline
- 說明：以當下時間進行離線收益結算（MVP）。
- 請求（可選參數，便於測試）：
```
{
  "asOf": "2025-08-14T10:00:00Z"   // RFC3339，可選；不帶則為 server Now()
}
```
- 回傳 200 JSON：
```
{
  "result": {
    "gainedKnowledge": 600,
    "gainedResearch": 120,
    "clampedTo8h": true,
    "anomalyDetected": false,
    "message": ""
  },
  "viewModel": {
    "knowledge": 12945,
    "research": 798,
    "notices": []
  }
}
```
- 型別對應：
  - `app/domain/gametime.OfflineResult`
  - `app/usecase/dto/game.ViewModelDto`

### POST /api/v1/game/start-practice
- 說明：立即開始一個練習任務（固定示範型）。
- 回傳：200 JSON，最新 ViewModel（含 currentTask）。

### POST /api/v1/game/try-finish
- 說明：嘗試完成當前任務；若尚未到時間，回傳 finished=false。
- 回傳 200 JSON：
```
{
  "finished": false,
  "reward": 0,
  "viewModel": { ... }
}
```

### （可選）POST /api/game/simulate-offline
- 說明：為了開發/測試方便，將儲存的 `timestamps.WallClockAtClose` 往前回推指定時長，以便立即看到離線收益。
- 請求：
```
{
  "back": "30m"   // Go duration 字串，如 30m、2h
}
```
- 回傳：200 OK（空物件或回傳最新 ViewModel）
- 實作注意：對 bbolt 與 memory 需統一流程：Load -> 調整 ts -> Save。

## 資料流與初始化
- Adapter 啟動時建立：
  - Repository（bbolt 或記憶體）
  - Clock（系統時鐘）
  - OfflineCalculator
  - Interactor（Usecase）並呼叫 `Initialize()`
- Handler 內持有 Usecase 例項，請求時進行對應呼叫。

## 錯誤處理與格式
- Content-Type: `application/json; charset=utf-8`
- 統一錯誤格式：
```
{
  "error": {
    "code": "bad_request|internal|not_found|...",
    "message": "..."
  }
}
```
- 主要錯誤情境：
  - 參數格式錯誤（400）
  - 內部錯誤（500）

## 佈署與開發
- 本地開發：在 `cmd/server` 下掛載 HTTP 路由：
  - HTTP 伺服器與路由位於 `app/adapter/httpserver`：
    - `handler.go`：依賴 port/in Usecase 與 DTO
    - `routes.go`：註冊 /api/v1/* 路徑，並保留 /api/* 相容
    - `middleware.go`：Request ID、Recovery、AccessLog
    - `server.go`：Start/Shutdown 包裝
- 與 fx 結合：在 `cmd/server.go` 的 `fx.New(...)` 中 Provide 必要元件與 Invoke 啟動 HTTP。
- 預設監聽：`127.0.0.1:8080`

## 安全性與限制
- 僅用於本地開發（loopback 介面）。若要跨裝置或產線：
  - 加上認證（token header）
  - CORS 設定
  - 速率限制

## 後續擴充
- 任務相關 API：開始任務、輪詢狀態、完成通知
- 學習/升級 API：升級語言等級、查看曲線與加成
- 事件/通知串流：SSE/WebSocket（本地 UI 可即時更新）
- 存讀檔：手動存檔/讀檔端點（或自動）

## 範例（可選）
- 取得 ViewModel：
```
GET http://127.0.0.1:8080/api/v1/game/viewmodel
```
- 結算離線（以現在）：
```
POST http://127.0.0.1:8080/api/v1/game/claim-offline
Content-Type: application/json

{}
```
- 結算離線（指定 asOf）：
```
POST http://127.0.0.1:8080/api/v1/game/claim-offline
Content-Type: application/json

{"asOf":"2025-08-14T10:00:00Z"}
```
- 開始練習：
```
POST http://127.0.0.1:8080/api/v1/game/start-practice
```
- 嘗試完成：
```
POST http://127.0.0.1:8080/api/v1/game/try-finish
```
- 模擬回推 30 分鐘：
```
POST http://127.0.0.1:8080/api/game/simulate-offline
Content-Type: application/json

{"back":"30m"}
```
