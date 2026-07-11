# Digital Rebound

**中文名稱：** 數位反彈效應分析  
**建築學來源：** Rebound Effect / Jevons Paradox（反彈效應）

## 核心問題

> 單位資源效率提升後，為何總成本、能耗或運算量沒有下降，甚至上升？

## Problem statement

最佳化常降低單位成本，卻刺激更多需求、更多部署或更鬆散的使用紀律，使總消耗反而增加。

## Inputs

- 最佳化事件
- 成本與能耗時間序列
- request/token/build volume
- 容量與採用率
- 產品發布紀錄


## Outputs

- Rebound Ratio
- Backfire alerts
- 節省被抵銷的來源分解
- 行為與容量成長關聯


## Primary metrics

| Metric ID | 初始用途 |
|---|---|
| `rebound_ratio` | 建立 baseline、趨勢與 review trigger；正式公式見後續 metric registry。 |
| `unit_cost_delta` | 建立 baseline、趨勢與 review trigger；正式公式見後續 metric registry。 |
| `total_consumption_delta` | 建立 baseline、趨勢與 review trigger；正式公式見後續 metric registry。 |
| `demand_elasticity` | 建立 baseline、趨勢與 review trigger；正式公式見後續 metric registry。 |
| `backfire_probability` | 建立 baseline、趨勢與 review trigger；正式公式見後續 metric registry。 |


## MVP scope

1. 標記最佳化事件
2. 建立前後視窗
3. 計算單位與總量變化
4. 產生反彈分級


## Out of scope for MVP

- 自動執行高風險變更
- 以單一分數取代專業審查
- 未經授權蒐集個人、員工、病患或客戶敏感資料
- 宣稱已建立普遍適用的因果關係

## Repository contract

- `project.yaml`：machine-readable project manifest
- `api/openapi.yaml`：最小 ingestion / assessment API
- `schemas/event.schema.json`：事件資料契約
- `docs/architecture.md`：邊界、資料流與 implementation slices
- `docs/threat-model.md`：安全、隱私與濫用風險
- `examples/sample-event.json`：合成資料範例

## First implementation milestone

先完成 **single-tenant, synthetic-data, read-only analysis**。在證據追溯、權限、資料保留與人工覆核尚未建立前，不進入真實組織或個人資料環境。
