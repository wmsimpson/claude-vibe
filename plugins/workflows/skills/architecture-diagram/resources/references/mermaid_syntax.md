# Mermaid Syntax Quick Reference

Quick reference for creating Mermaid diagrams.

---

## Flowchart (Most Common)

```mermaid
flowchart LR
    A[Rectangle] --> B(Rounded)
    B --> C{Decision}
    C -->|Yes| D[Result 1]
    C -->|No| E[Result 2]
```

### Direction
- `TB` / `TD` - Top to bottom
- `BT` - Bottom to top
- `LR` - Left to right
- `RL` - Right to left

### Node Shapes
```
A[Rectangle]
B(Rounded rectangle)
C([Stadium/pill])
D[[Subroutine]]
E[(Database)]
F((Circle))
G>Asymmetric]
H{Diamond/decision}
I{{Hexagon}}
J[/Parallelogram/]
K[\Parallelogram alt\]
L[/Trapezoid\]
M[\Trapezoid alt/]
```

### Links/Arrows
```
A --> B          Solid arrow
A --- B          Solid line (no arrow)
A -.-> B         Dotted arrow
A -.- B          Dotted line
A ==> B          Thick arrow
A === B          Thick line
A --text--> B    Arrow with text
A ---|text|B     Line with text
A -->|text| B    Arrow with text (alt)
```

### Subgraphs
```mermaid
flowchart LR
    subgraph Sources["Data Sources"]
        A[Kafka]
        B[S3]
    end

    subgraph Processing
        C[ETL]
    end

    A --> C
    B --> C
```

---

## Sequence Diagram

```mermaid
sequenceDiagram
    autonumber
    participant A as Client
    participant B as Server
    participant C as Database

    A->>B: Request
    B->>C: Query
    C-->>B: Results
    B-->>A: Response

    Note over A,B: This is a note
    Note right of C: Database note
```

### Arrow Types
```
A->B     Solid line
A-->B    Dotted line
A->>B    Solid with arrowhead
A-->>B   Dotted with arrowhead
A-xB     Solid with X
A--xB    Dotted with X
A-)B     Solid with open arrow
A--)B    Dotted with open arrow
```

### Activation
```mermaid
sequenceDiagram
    A->>+B: Request
    B->>+C: Query
    C-->>-B: Results
    B-->>-A: Response
```

### Loops and Conditionals
```mermaid
sequenceDiagram
    loop Every minute
        A->>B: Poll
    end

    alt Success
        B-->>A: Data
    else Error
        B-->>A: Error
    end
```

---

## Entity Relationship Diagram

```mermaid
erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains
    PRODUCT ||--o{ LINE-ITEM : "is in"

    CUSTOMER {
        int id PK
        string name
        string email
    }

    ORDER {
        int id PK
        int customer_id FK
        date created
    }
```

### Cardinality
```
||--||   One to one
||--o{   One to many
o{--o{   Many to many
||--o|   One to zero or one
```

---

## State Diagram

```mermaid
stateDiagram-v2
    [*] --> Draft
    Draft --> Review
    Review --> Approved: Approve
    Review --> Draft: Reject
    Approved --> Published
    Published --> [*]
```

---

## Gantt Chart

```mermaid
gantt
    title Project Timeline
    dateFormat  YYYY-MM-DD

    section Planning
    Requirements     :a1, 2024-01-01, 30d
    Design          :a2, after a1, 20d

    section Development
    Sprint 1        :b1, after a2, 14d
    Sprint 2        :b2, after b1, 14d

    section Testing
    QA              :c1, after b2, 10d
```

---

## Class Diagram

```mermaid
classDiagram
    class Animal {
        +String name
        +int age
        +makeSound()
    }

    class Dog {
        +fetch()
    }

    Animal <|-- Dog
```

---

## Pie Chart

```mermaid
pie title Data Distribution
    "Bronze" : 40
    "Silver" : 35
    "Gold" : 25
```

---

## Theming

```mermaid
%%{init: {
    'theme': 'base',
    'themeVariables': {
        'primaryColor': '#FF3621',
        'primaryTextColor': '#fff',
        'primaryBorderColor': '#FF3621',
        'lineColor': '#333',
        'secondaryColor': '#f5f5f5',
        'tertiaryColor': '#fff'
    }
}}%%
flowchart LR
    A --> B --> C
```

### Built-in Themes
- `default`
- `base`
- `dark`
- `forest`
- `neutral`

---

## Generating Output

### Using Mermaid CLI (mmdc)

```bash
# Install
npm install -g @mermaid-js/mermaid-cli

# Generate PNG
mmdc -i diagram.mmd -o diagram.png

# Generate SVG
mmdc -i diagram.mmd -o diagram.svg

# Generate PDF
mmdc -i diagram.mmd -o diagram.pdf

# With config
mmdc -i diagram.mmd -o diagram.png -c config.json

# With background color
mmdc -i diagram.mmd -o diagram.png -b white
```

### Config File (config.json)

```json
{
  "theme": "base",
  "themeVariables": {
    "primaryColor": "#FF3621"
  }
}
```

---

## draw.io Import

1. Open draw.io
2. File > Import From > Text
3. Select "Mermaid" format
4. Paste your .mmd content
5. Click "Insert"

The diagram will be imported as editable shapes.

---

## Live Editors

- [Mermaid Live Editor](https://mermaid.live/)
- [draw.io](https://app.diagrams.net/) (native Mermaid support)
- VS Code with Mermaid extension

---

## Tips

1. **Keep it simple**: Mermaid is best for simpler diagrams
2. **Use subgraphs**: Group related components
3. **Direction matters**: `LR` for wide, `TB` for tall
4. **Labels**: Use `|text|` for edge labels
5. **Styling**: Use `%%{init:...}%%` for theming
