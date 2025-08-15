# Ebiten × 後端 Adapter 串接設計（外部 App 方案）

本文件說明：將 Ebiten 作為「外部桌面 App」，不直接內嵌於後端程式碼；透過後端 Adapter（HTTP/JSON 為主，日後可升級 gRPC/WebSocket）與 Usecase Port 串接，達到清晰分層與良好延展性。

## 1) 定位與原則

- Ebiten：僅負責輸入/渲染與畫面狀態管理（Presenter/UI），不實作業務規則。
- 後端（本專案）：
  - Domain：純邏輯（Player/Resource/Gametime/...），不依賴外部技術。
  - Usecase：以 Port（Input/Output）協調流程，對 UI 暴露意圖型呼叫與 ViewModel。
  - Infra：技術細節（bbolt/Clock/Log）。
  - Adapter（對外）：HTTP/gRPC/WebSocket，將 Usecase 能力以 API 形式提供。
- 資料流向：UI(App) → Adapter(HTTP) → Usecase(Port-In) → Domain；Usecase(Port-Out) → Infra（Repository/Clock）。
- 時間權威：由後端提供 Clock；UI 不直接上傳 now（避免時間作弊）。

## 2) 後端目錄規劃（Adapter 與既有分層）

```
app/
  adapter/
    transport/
      http/
        router.go          # 註冊路由、middleware
        game/
          controller.go    # HTTP handler → Usecase 呼叫
          dto.go           # HTTP 請/回 DTO（JSON）
  usecase/
    port/
      in/game/game.go      # 已存在：Input Port（Usecase 介面）
      out/game/repository.go # 已存在：Output Port（Repository 介面）
    game/
      interactor.go        # 已存在：協調者（Usecase 實作）
      dto/viewmodel.go     # 已存在：Usecase → Presenter 的 ViewModel
  domain/
    gametime/ ...          # 已存在
    player/ ...            # 已存在
    resource/ ...          # 已存在
  infra/
    persistence/bbolt/ ... # MVP 將新增（Repository 實作）
    memory/ ...            # 已存在（InMemory 用於測試/展示）
    clock/  ...            # 已存在（SystemClock）
cmd/
  server.go                # 啟動 HTTP Server、DI wiring
```

## 3) API 契約（MVP 草案）

以 HTTP+JSON 起步，簡單易除錯；日後可補 Protobuf/gRPC 與 WebSocket/SSE。

- POST `/v1/game/init`
  - 動作：載入狀態（Usecase.Initialize）並執行一次離線收益（Usecase.ClaimOffline，採後端時間）。
  - 回應：{ ok: bool, notices: [string] }

- POST `/v1/game/claim-offline`
  - 動作：強制執行離線收益（通常啟動後自動，不需頻繁呼叫）。
  - 回應：{ gainedKnowledge: int64, gainedResearch: int64, clampedTo8h: bool, anomalyDetected: bool, message?: string }

- GET `/v1/game/view-model`
  - 動作：取得當前 ViewModel（資源、提示、練習/任務狀態等）。
  - 回應：{ resources: { knowledge, research }, notices: [string], practice?: { language, level, xp, remainingSeconds? } }

-（M2+）POST `/v1/game/start-practice`
  - body：{ language: string, durationSeconds: int }
  - 動作：開始練習，後端內部以「開始時間 + 速度」做函式性進度推算。

-（M2+）POST `/v1/game/start-solve`
  - body：{ taskId: string }

錯誤策略：
- 4xx：參數/狀態錯誤（例如正在練習時再度開始）。
- 5xx：伺服器錯誤（例如持久化失敗）。
- 回應建議含 { code, message }，便於 UI 顯示。

## 4) 啟動與主迴圈（UI 側）

1. Ebiten App 啟動：
   - 呼叫 POST /v1/game/init → 後端載入 + ClaimOffline。
   - 呼叫 GET /v1/game/view-model → 若有離線收益，Presenter 顯示提示彈窗。
2. Update/Draw 主迴圈：
   - Update：處理鍵鼠輸入，對應動作發送 POST（StartPractice/Solve/...）。
   - ViewModel 取得：每 0.5～1 秒輪詢 GET /view-model（或升級 WebSocket/SSE 推送）。
   - Draw：依 ViewModel 與 UI 狀態渲染。
3. 結束：後端會自行背景保存（去抖動），無需 UI 強制同步等待。

## 5) 時間與平行處理（Concurrency）策略

- 時間權威：後端的 Clock.Now()。
- 離線收益：在 init/claim 時於後端計算，UI 僅展示結果。
- 保存：後端以背景 goroutine + 佇列去抖動保存，避免阻塞任何 UI 或 HTTP handler。
- 資料競態：Usecase 持有聚合根；保存使用「快照」，避免與主邏輯共享可變狀態。

## 6) 測試與品質

- Domain/Usecase：單元/端對端測試（已建立離線收益與 8 小時上限）。
- Adapter(HTTP)：使用 `httptest` 做 handler 測試，含序列化、錯誤碼與路由。
- 契約一致性：可附上 OpenAPI（可選），或在前端 repo 放小型 client + smoke 測試。
- 品質關卡：`go test ./...` 全綠、無循環相依、文件與 README 同步。

## 7) MVP 路線圖（與後續）

- M1（後端）：
  - bbolt Repository 實作（Game Repository）。
  - HTTP Adapter：`/v1/game/init`、`/v1/game/claim-offline`、`/v1/game/view-model`。
  - CLI smoke 維持（除錯友善）。
- M2：
  - Practice/Advance/StartSolve 用例與 API。
  - ViewModel 擴充（練習狀態、剩餘時間、語言進度）。
- M3：
  - 時間異常完整化（單調時間代理值門檻→保守結算）、UI 提示、日誌觀測。
  - 若需要即時提示，升級 WebSocket/SSE 推送。

## 8) 風險與決策點

- 單調時間代理值：Go 的 monotonic 無法跨重啟，需自存代理值（ElapsedMonotonicSeconds）或近似指標。
- 儲存方案：MVP 用 bbolt（單檔、ACID），mongo 留作擴展。
- UI 輪詢 vs 推播：MVP 用輪詢（0.5～1s），之後再決定是否引入長連線。
- 打包體驗：Ebiten 可檢查後端是否啟動；若無，提示使用者或嘗試啟動子行程（可選）。

---

以上設計確保 Ebiten 與後端解耦、契約清晰、可測可演進。後續若決定端點/欄位細節，我們可以在此檔追加簡易 OpenAPI 草案以利前後端平行開發。
