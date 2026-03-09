# Lean Process Map Symbols & Patterns Reference

Standard symbols used in Lean Six Sigma process maps, with Mermaid syntax and Graphviz DOT equivalents for generating diagrams programmatically.

---

## Standard Symbols

| Symbol | Name | Mermaid Syntax | DOT Shape | When to Use |
|--------|------|---------------|-----------|-------------|
| Oval | Start / End | `([Text])` | `shape=oval` | Process start trigger, terminal state |
| Rectangle | Process Step | `[Text]` | `shape=box` | Any activity or task |
| Diamond | Decision | `{Text}` | `shape=diamond` | Yes/No branch, conditional logic |
| Parallelogram | Input/Output | `/[Text]/` | `shape=parallelogram` | Document, data, or artifact |
| Cylinder | Data Store | `[(Text)]` | `shape=cylinder` | Database, Delta table, file system |
| Rounded Rectangle | External Entity | — | `shape=box, style=rounded` | Customer, external system, partner |
| Wave-bottom Rectangle | Delay/Wait | `[⏳ Wait]` | `shape=box` with NVA color | Waiting state — always NVA |
| Arrow → | Flow | `-->` | `->` | Sequential flow |
| Dashed Arrow | Optional/Rework | `-.->` | `[style=dashed]` | Rework loop, exception path |
| Bold Arrow | Critical Path | `==>` | `[penwidth=3]` | Main happy path |

---

## Value Classification Colors

| Classification | Hex Color | Meaning |
|----------------|-----------|---------|
| VA (Value-Added) | `#C8E6C9` / `#2E7D32` | Customer pays for this; keep and optimize |
| NVA (Waste) | `#FFCDD2` / `#C62828` | Eliminate; target for kaizen |
| NNVA (Necessary Non-VA) | `#FFF9C4` / `#F9A825` | Minimize; automate where possible |
| Decision | `#E3F2FD` / `#1565C0` | Conditional branch |
| Data Store | `#EDE7F6` / `#6A1B9A` | Storage artifact |
| Start/End | `#E8F5E9` / `#2E7D32` | Process boundary |

---

## Swimlane Color Palette

Assign one background color per swimlane (role/team):

| Role Type | Suggested Color | Hex |
|-----------|----------------|-----|
| Data Engineering | Light blue | `#E3F2FD` |
| Data Science / ML | Light purple | `#F3E5F5` |
| Data Governance / Ops | Light amber | `#FFF8E1` |
| Business / Analyst | Light green | `#E8F5E9` |
| External / Customer | Light red | `#FFEBEE` |
| Platform / Infrastructure | Light gray | `#ECEFF1` |
| Security / Compliance | Light orange | `#FBE9E7` |

---

## Mermaid Patterns

### Basic Process Flow (no swimlanes)

```mermaid
flowchart TD
    A([▶ Start: New data arrives]) --> B["Validate schema\n(VA)"]:::va
    B --> C{Schema valid?}:::decision
    C -->|Yes| D["Apply transformations\n(VA)"]:::va
    C -->|No| E["Quarantine record\n(NNVA)"]:::nnva
    E --> F["Alert data team\n(NNVA)"]:::nnva
    F --> G([⏹ End: Record quarantined])
    D --> H[("Write to Silver table\n(VA)")]:::va
    H --> G2([⏹ End: Silver updated])

    classDef va fill:#C8E6C9,stroke:#2E7D32,color:#1B5E20
    classDef nva fill:#FFCDD2,stroke:#C62828,color:#B71C1C
    classDef nnva fill:#FFF9C4,stroke:#F9A825,color:#F57F17
    classDef decision fill:#E3F2FD,stroke:#1565C0,color:#0D47A1
```

### Swimlane Flow (subgraph pattern)

```mermaid
flowchart LR
    START([▶ Trigger])

    subgraph DE["🏊 Data Engineering"]
        DE1["Ingest raw data\n(VA)"]:::va
        DE2["Apply DLT expectations\n(VA)"]:::va
        DE3{"Data quality\ngate?"}
    end

    subgraph GOV["🏊 Data Governance"]
        GOV1["Review quarantined\nrecords\n(NNVA)"]:::nnva
        GOV2["Approve remediation\n(NNVA)"]:::nnva
    end

    subgraph ANA["🏊 Analytics"]
        ANA1["Build Gold aggregates\n(VA)"]:::va
        ANA2["Publish to dashboard\n(VA)"]:::va
    end

    END_NODE([⏹ Data available])

    START --> DE1
    DE1 --> DE2
    DE2 --> DE3
    DE3 -->|Pass| ANA1
    DE3 -->|Fail| GOV1
    GOV1 --> GOV2
    GOV2 -->|Remediated| DE2
    ANA1 --> ANA2
    ANA2 --> END_NODE

    classDef va fill:#C8E6C9,stroke:#2E7D32,color:#1B5E20
    classDef nnva fill:#FFF9C4,stroke:#F9A825,color:#F57F17
```

---

## Graphviz DOT Patterns

### Complete Swimlane Template

```dot
digraph ProcessMap {
    rankdir=TB;
    compound=true;
    newrank=true;
    node [fontname="Arial", fontsize=10, style="filled,rounded"];
    edge [color="#555555", penwidth=1.5, fontname="Arial", fontsize=9];
    graph [splines=polyline, nodesep=0.6, ranksep=1.0, pad=0.5,
           fontname="Arial Bold", fontsize=14];

    // ── Process boundary markers ──
    START [label="START\nTrigger Event", shape=oval,
           fillcolor="#E8F5E9", color="#2E7D32", penwidth=2, fontname="Arial Bold"];
    END   [label="END\nProcess Complete", shape=oval,
           fillcolor="#E8F5E9", color="#2E7D32", penwidth=2, fontname="Arial Bold"];

    // ── Swimlane 1: Data Engineering ──
    subgraph cluster_de {
        label="Data Engineering";
        style=filled; fillcolor="#E3F2FD"; color="#1565C0"; penwidth=2;
        fontname="Arial Bold"; fontsize=12;

        DE1 [label="1. Ingest Events\n[VA]", shape=box,
             fillcolor="#C8E6C9", color="#2E7D32"];
        DE2 [label="2. Validate Schema\n[VA]", shape=box,
             fillcolor="#C8E6C9", color="#2E7D32"];
        DE_DEC [label="Schema\nValid?", shape=diamond,
                style=filled, fillcolor="#E3F2FD", color="#1565C0"];
        DE3 [label="3. Transform\nBronze→Silver\n[VA]", shape=box,
             fillcolor="#C8E6C9", color="#2E7D32"];
        DE_DB [label="Silver\nDelta Table", shape=cylinder,
               style=filled, fillcolor="#EDE7F6", color="#6A1B9A"];
    }

    // ── Swimlane 2: Data Governance ──
    subgraph cluster_gov {
        label="Data Governance";
        style=filled; fillcolor="#FFF8E1"; color="#F57F17"; penwidth=2;
        fontname="Arial Bold"; fontsize=12;

        GOV1 [label="Review Quarantine\n[NNVA]", shape=box,
              fillcolor="#FFF9C4", color="#F9A825"];
        GOV2 [label="Approve &\nRemediate\n[NNVA]", shape=box,
              fillcolor="#FFF9C4", color="#F9A825"];
    }

    // ── Swimlane 3: Analytics ──
    subgraph cluster_ana {
        label="Analytics";
        style=filled; fillcolor="#E8F5E9"; color="#2E7D32"; penwidth=2;
        fontname="Arial Bold"; fontsize=12;

        ANA1 [label="4. Aggregate\nSilver→Gold\n[VA]", shape=box,
              fillcolor="#C8E6C9", color="#2E7D32"];
        ANA_OUT [label="Gold\nDashboard", shape=parallelogram,
                 style=filled, fillcolor="#E8EAF6", color="#283593"];
    }

    // ── Flows ──
    START -> DE1;
    DE1 -> DE2;
    DE2 -> DE_DEC;
    DE_DEC -> DE3 [label="Pass"];
    DE_DEC -> GOV1 [label="Fail", style=dashed, color="#C62828"];
    GOV1 -> GOV2;
    GOV2 -> DE2 [label="Retry", style=dashed, color="#F57F17"];
    DE3 -> DE_DB;
    DE_DB -> ANA1 [lhead=cluster_ana];
    ANA1 -> ANA_OUT;
    ANA_OUT -> END;
}
```

---

## Common Databricks Process Patterns

### Bronze → Silver → Gold Pipeline

```mermaid
flowchart LR
    subgraph SRC["📡 Sources"]
        S1["Kafka\nStream"]
        S2["S3\nFiles"]
    end

    subgraph DBX["🏠 Databricks Lakehouse"]
        subgraph DE["🏊 Data Engineering"]
            B["⬛ Bronze\nRaw Ingest\n(VA)"]:::va
            S["🥈 Silver\nValidate+Clean\n(VA)"]:::va
            G["🥇 Gold\nAggregate\n(VA)"]:::va
            DQ{"DQ\nGate?"}
            Q["Quarantine\n(NNVA)"]:::nnva
        end

        subgraph GOV["🏊 Governance"]
            UC["Unity Catalog\nRegister+Tag\n(NNVA)"]:::nnva
            LIN["Lineage\nCapture\n(NNVA)"]:::nnva
        end
    end

    subgraph CONS["👥 Consumers"]
        BI["📊 BI\nDashboard"]
        DS["🤖 Data\nScience"]
        EXT["🤝 External\nDelta Share"]
    end

    S1 --> B
    S2 --> B
    B --> DQ
    DQ -->|Pass| S
    DQ -->|Fail| Q
    S --> G
    G --> UC
    G --> LIN
    UC --> BI
    UC --> DS
    UC --> EXT

    classDef va fill:#C8E6C9,stroke:#2E7D32,color:#1B5E20
    classDef nnva fill:#FFF9C4,stroke:#F9A825,color:#F57F17
```

### ML Pipeline (Training → Serving)

```mermaid
flowchart TD
    subgraph DS_TEAM["🏊 Data Science"]
        FE["Feature\nEngineering\n(VA)"]:::va
        TRAIN["Model\nTraining\n(VA)"]:::va
        EVAL{"Meets\naccuracy\nthreshold?"}
    end

    subgraph MLOPS["🏊 MLOps / Platform"]
        REG["Register in\nMLflow\n(NNVA)"]:::nnva
        REVIEW["Manual\nReview\n(NNVA)"]:::nnva
        DEPLOY["Deploy to\nModel Serving\n(VA)"]:::va
        MON["Monitor\nDrift\n(NNVA)"]:::nnva
    end

    subgraph APP["🏊 Application"]
        INF["Real-time\nInference\n(VA)"]:::va
        OUT[/"Prediction\nResponse"/]
    end

    FE --> TRAIN
    TRAIN --> EVAL
    EVAL -->|Pass| REG
    EVAL -->|Fail| FE
    REG --> REVIEW
    REVIEW --> DEPLOY
    DEPLOY --> MON
    MON -->|Drift| FE
    DEPLOY --> INF
    INF --> OUT

    classDef va fill:#C8E6C9,stroke:#2E7D32,color:#1B5E20
    classDef nnva fill:#FFF9C4,stroke:#F9A825,color:#F57F17
```

---

## Cycle Time Annotation

Add cycle time data to each step for Value Stream Mapping:

```mermaid
flowchart LR
    S1["Step 1\n⏱ PT: 5 min\n⌛ WT: 2 hrs"]:::va --> S2
    S2["Step 2\n⏱ PT: 30 min\n⌛ WT: 4 hrs (NVA!)"]:::nnva --> S3
    S3["Step 3\n⏱ PT: 10 min\n⌛ WT: 0"]:::va
```

`PT` = Process Time (actual work)
`WT` = Wait Time (idle, queuing — often NVA)

**Process Efficiency = Total PT / (Total PT + Total WT)**
- Target: > 60% for a lean process
- Typical: 5–20% in non-optimized processes (80%+ is waste)
