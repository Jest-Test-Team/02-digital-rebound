# Implementation Plan — Digital Rebound

Status: in progress (Phase 1 MVP)  
Aligned with: [`README.md`](../../README.md), [`docs/architecture.md`](../architecture.md), [`docs/threat-model.md`](../threat-model.md), [`project.yaml`](../../project.yaml)  
Last updated: 2026-07-12

## 1. Purpose

將 blueprint 推進到 **single-tenant, synthetic-data, read-only analysis** MVP，垂直切片：

```text
ingest → validate → calculate → explain → review
```

核心問題：單位資源效率提升後，為何總成本、能耗或運算量沒有下降，甚至上升？

每個 deliverable 必須維持治理約束：

- 僅合成 / 非生產資料
- evidence provenance 必填
- 所有 assessment 的 `human_review_required = true`
- 不以單一分數取代專業審查
- 不宣稱已建立普遍適用的因果關係

## 2. Current baseline

| Area | Present today | Gap |
| --- | --- | --- |
| Manifest | `project.yaml` | status 仍為 `concept` / `blueprint` |
| API contract | `api/openapi.yaml`（events POST、assessments GET） | 無實作 server |
| Event schema | `schemas/event.schema.json` | 無 runtime validation |
| Synthetic fixture | `examples/sample-event.json` | 缺前後視窗 / 反例 / 缺失資料案例 |
| Source | `src/README.md` placeholder | 無 Go/Gin 服務 |
| Tests | `tests/README.md` 涵蓋清單 | 無 Robot Framework suite |
| CI / release | 無 | 需 GHA + Docker image release |

## 3. Architecture targets (Phase 1)

```text
Synthetic fixtures / importers
        ↓
Schema validation (JSON Schema 0.1.0)
        ↓
Evidence store (file-based, append-only)
        ↓
Metric engine (explainable formulas, pinned rule version)
        ↓
ReboundAssessment + human-review gate
        ↓
Review API (OpenAPI 0.1.0)
```

Scale path（不得跳級）：

1. Phase 1: file-based fixtures + batch analysis ← **本計畫範圍**
2. Phase 2: PostgreSQL + object evidence + scheduled jobs
3. Phase 3: graph/time-series + streaming ingestion
4. Phase 4: multi-tenant policy + audit + model/rule registry

### Non-goals (MVP)

- 自動執行高風險變更
- ML / 黑箱模型
- 真實組織或個人資料
- 跨租戶生產授權
- 以單一分數取代人工覆核

## 4. Domain & metrics

### 4.1 Entities

| Entity | Role |
| --- | --- |
| `OptimizationEvent` | 標記效率最佳化發生點 |
| `ConsumptionSeries` | 成本 / 能耗時間序列 |
| `DemandSeries` | request / token / build volume |
| `Baseline` | 前後視窗比較基準 |
| `ReboundAssessment` | 可解釋評估結果 + review 狀態 |

### 4.2 Primary metrics（可解釋、無 ML）

| Metric ID | MVP 公式（pinned `rule_version`） | 用途 |
| --- | --- | --- |
| `unit_cost_delta` | `(unit_after - unit_before) / unit_before` | 單位成本變化 |
| `total_consumption_delta` | `(total_after - total_before) / total_before` | 總消耗變化 |
| `rebound_ratio` | `1 - (actual_savings / expected_savings)`；`expected_savings = -unit_cost_delta * baseline_total` | 節省被抵銷比例 |
| `demand_elasticity` | `total_consumption_delta / unit_cost_delta`（unit_cost_delta ≠ 0） | 需求對效率的彈性 |
| `backfire_probability` | 規則分數：rebound_ratio ≥ 1 → 1.0；≥ 0.5 → 0.7；≥ 0.2 → 0.4；否則 0.1（含缺失資料懲罰） | backfire 警示強度 |

### 4.3 Rebound grading

| Grade | Condition |
| --- | --- |
| `none` | `rebound_ratio ≤ 0` |
| `partial` | `0 < rebound_ratio < 1` |
| `backfire` | `rebound_ratio ≥ 1` |
| `inconclusive` | 缺失資料 / 視窗不足 / 預期節省 ≈ 0 |

## 5. Phased work

### 5.1 Slice A — Contract & fixtures（前置）

| ID | Work item | Acceptance |
| --- | --- | --- |
| A-1 | 維持 `schemas/event.schema.json`、`api/openapi.yaml` 為單一真相來源 | CI schema 檢查通過 |
| A-2 | 擴充合成 fixtures：happy path、missing data、stale evidence、counterexample、cross-tenant | 檔案置於 `examples/` 與 `tests/fixtures/` |
| A-3 | 文件化 `rule_version`（例如 `rebound-rules@0.1.0`） | assessment 回應含 pinned version |

**Exit:** 契約與 fixtures 可獨立驗證，不依賴 server。

---

### 5.2 Slice B — Collection（ingest）

| ID | Work item | Acceptance |
| --- | --- | --- |
| B-1 | Go + Gin server：`POST /v1/digital-rebound/events` | 符合 OpenAPI；成功回 `202` |
| B-2 | Runtime JSON Schema validation | 缺欄位 / 錯 metric / 錯 schema_version → `400` |
| B-3 | 寫入 evidence store，保留 `source`、`observed_at`、`ingested_at`、content hash | append-only；可追溯 |
| B-4 | Tenant scope：store key 含 `tenant_id` | 跨租戶讀取被隔離 |

**Exit:** 可 ingest 合成事件並拒絕非法 payload。

---

### 5.3 Slice C — Analysis（calculate + explain）

| ID | Work item | Acceptance |
| --- | --- | --- |
| C-1 | 前後視窗 baseline 建構（event 前後 N 點或時間窗） | 視窗不足 → `inconclusive` |
| C-2 | 實作 §4.2 五個 metric（至少 `rebound_ratio` + `unit_cost_delta` + `total_consumption_delta` 完整） | 單元測試覆蓋公式邊界 |
| C-3 | Assessment 輸出含 evidence_refs、assumptions、missing_data、uncertainty、rule_version | 可解釋、可覆核 |
| C-4 | Backfire alert 僅為 review candidate，非自動動作 | `human_review_required=true` 常數 |

**Exit:** 對合成案例產生分級與解釋，無黑箱。

---

### 5.4 Slice D — Review API

| ID | Work item | Acceptance |
| --- | --- | --- |
| D-1 | `GET /v1/digital-rebound/assessments` | 回傳 array of Assessment |
| D-2 | Assessment status ∈ `draft` \| `review-required` \| `accepted` \| `rejected` | 預設 `review-required` |
| D-3 | `GET /healthz` | CI / Robot smoke 使用 |
| D-4 | （可選）`POST .../assessments/{id}/annotations` append-only | 不覆寫 evidence |

**Exit:** Reviewer 可列出待審評估與 evidence 引用。

---

### 5.5 Slice E — Evaluation（Robot Framework）

涵蓋 `tests/README.md` 要求：

| ID | Suite / 案例 | Acceptance |
| --- | --- | --- |
| E-1 | Schema validation | 合法 sample → 202；缺必填 / 非法 metric → 400 |
| E-2 | Missing data | 缺 series → assessment `inconclusive` + missing_data 非空 |
| E-3 | Stale evidence | `observed_at` 過舊相對 ingest → uncertainty 升高或標註 stale |
| E-4 | Counterexample | 效率↑且總量↓ → grade `none`（非 backfire） |
| E-5 | Tenant isolation | tenant-A 不可見 tenant-B assessments / events |
| E-6 | Rule versioning | assessment 含固定 `rule_version` |
| E-7 | Human-review gate | 所有 assessment `human_review_required == true` |
| E-8 | Smoke health | `/healthz` → 200 |

測試布局：

```text
tests/
  README.md
  robot/
    requirements.txt
    resources/
      keywords.robot
      variables.robot
    fixtures/
      valid_event.json
      missing_fields.json
      stale_event.json
      counterexample_series.json
    suites/
      01_schema_validation.robot
      02_missing_data.robot
      03_stale_evidence.robot
      04_counterexample.robot
      05_tenant_isolation.robot
      06_rule_versioning.robot
      07_human_review_gate.robot
      08_health.robot
```

**Exit:** `make test-robot`（或等價腳本）在本機與 CI 全綠。

---

### 5.6 Slice F — Container & GHA

| ID | Work item | Acceptance |
| --- | --- | --- |
| F-1 | 精簡 `Dockerfile`（multi-stage Go build） | `docker build` 成功 |
| F-2 | `.github/workflows/ci.yml`：go test + robot | PR / push 必過 |
| F-3 | `.github/workflows/release-docker.yml`：tag `v*` 或 workflow_dispatch → GHCR | 推送 `ghcr.io/<owner>/<repo>/digital-rebound-api:<tag>` |

**Exit:** 合併門檻有自動化測試；釋出可重現映像。

## 6. Suggested repository layout (Phase 1)

```text
src/
  cmd/server/main.go
  internal/
    config/
    dto/          # request binding + validator
    vo/           # response / swagger-oriented
    handler/
    service/      # validate, metrics, assess
    store/        # file-based evidence + assessments
    rules/        # pinned rebound-rules@0.1.0
  go.mod
docs/dev-plan/implementation.plan.md
tests/robot/...
Dockerfile
Makefile
.github/workflows/ci.yml
.github/workflows/release-docker.yml
```

約束（對齊專案工程規範）：

- Handler / service 不直接暴露內部 store struct 當 API 契約；使用 DTO / VO
- Model 僅在引入 Postgres（Phase 2）時以 GORM struct 對應 schema
- Migration 僅在 Phase 2 新增於 `database/migrations/`（不可改舊檔）

## 7. Security & threat-model mapping

| Threat | MVP control |
| --- | --- |
| Fabricated evidence | schema + source identity + content hash |
| Cross-tenant access | tenant-scoped store keys；Robot E-5 |
| Unauthorized inference | synthetic-only；無個人資料欄位 |
| Recommendation tampering | human-review gate；無自動執行動作 |
| Stale topology as truth | stale evidence 標註 + uncertainty |
| 忽略季節性 / 合理成長 | counterexample + inconclusive；文件聲明非因果 |

## 8. Definition of done (per sub-task)

1. 對應 slice acceptance 通過
2. Robot 或單元測試覆蓋 happy path + 至少一條拒絕路徑
3. Docs / OpenAPI / schema 若有契約變更則同步
4. 子任務完成後 **立即 commit 並 push**（本計畫執行約定）
5. 不提交 secrets、真實租戶資料或 `.env`

## 9. Implementation order (commit cadence)

```text
1. docs/dev-plan/implementation.plan.md     → commit + push
2. Go MVP (health, ingest, assess, metrics) → commit + push
3. Robot Framework suites + fixtures          → commit + push
4. Verify tests; fix failures                 → commit + push
5. Dockerfile + GHA ci/release-docker         → commit + push
```

## 10. Risks

| Risk | Mitigation |
| --- | --- |
| 公式過度簡化被誤用為因果 | assessment 固定 disclaimer + review-required |
| 視窗參數影響分級 | pin `rule_version`；文件化視窗假設 |
| CI 無服務可測 | job 內建置並啟動 server，再跑 robot |
| Docker 映像膨脹 | multi-stage；distroless/alpine；無多餘套件 |

## 11. Exit criteria — Phase 1 MVP

- [ ] `POST /v1/digital-rebound/events` 與 `GET /v1/digital-rebound/assessments` 可用
- [ ] 合成資料可產生 rebound grade 與 explainable metrics
- [ ] Robot suites E-1…E-8 全過
- [ ] CI 綠燈；release workflow 可推 GHCR image
- [ ] `project.yaml` 可更新為 `status: mvp` / `maturity: phase-1`（於 MVP 驗收後）
