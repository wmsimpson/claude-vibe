# Graphviz Spacing & Layout Guide

Fine-tune diagram appearance with Graphviz attributes.

---

## Graph Attributes

Set these in the `Diagram()` constructor:

```python
with Diagram("Title",
             graph_attr={
                 # Spacing
                 "nodesep": "1.0",       # Horizontal space between nodes
                 "ranksep": "1.5",       # Vertical space between ranks
                 "pad": "0.5",           # Padding around entire diagram

                 # Edge routing
                 "splines": "ortho",     # Edge style (see below)

                 # Visual
                 "bgcolor": "white",     # Background color
                 "fontsize": "14",       # Default font size
                 "dpi": "150",           # Resolution

                 # Layout
                 "rankdir": "LR",        # Direction (alternative to direction param)
                 "compound": "true",     # Allow edges between clusters
             }):
```

---

## Edge Spline Styles

| Value | Description | Best For |
|-------|-------------|----------|
| `"ortho"` | 90-degree angles only | Clean technical diagrams |
| `"polyline"` | Straight segments | Simple layouts |
| `"curved"` | Smooth curves | Organic flow |
| `"spline"` | Bezier curves (default) | General use |
| `"line"` | Straight lines | Minimal diagrams |
| `"false"` or `"none"` | No routing | Debugging |

```python
graph_attr={"splines": "ortho"}  # Most professional look
```

---

## Direction Options

Set flow direction of the diagram:

| Value | Description |
|-------|-------------|
| `"TB"` | Top to bottom (default) |
| `"BT"` | Bottom to top |
| `"LR"` | Left to right |
| `"RL"` | Right to left |

```python
with Diagram("Title", direction="LR"):  # Horizontal flow
```

---

## Node Attributes

```python
with Diagram("Title",
             node_attr={
                 "fontsize": "12",      # Label font size
                 "width": "1.5",        # Minimum node width
                 "height": "1.5",       # Minimum node height
                 "fixedsize": "false",  # Allow nodes to grow
             }):
```

---

## Edge Attributes

```python
with Diagram("Title",
             edge_attr={
                 "fontsize": "10",      # Edge label font size
                 "penwidth": "2.0",     # Edge line thickness
                 "arrowsize": "1.0",    # Arrow head size
             }):
```

---

## Common Layout Problems & Solutions

### Problem: Nodes too close together

**Solution**: Increase spacing

```python
graph_attr={
    "nodesep": "1.5",    # Was 1.0
    "ranksep": "2.0",    # Was 1.5
}
```

### Problem: Edges crossing unnecessarily

**Solution 1**: Use orthogonal splines
```python
graph_attr={"splines": "ortho"}
```

**Solution 2**: Reorder nodes in code (order affects layout)
```python
# Instead of: a, c, b
# Use: a, b, c
```

**Solution 3**: Change direction
```python
direction="LR"  # Instead of "TB"
```

### Problem: Diagram too wide or tall

**Solution 1**: Change direction
```python
direction="TB"  # Vertical (narrower)
direction="LR"  # Horizontal (shorter)
```

**Solution 2**: Split into multiple clusters
```python
with Cluster("Left Side"):
    # Components here
with Cluster("Right Side"):
    # Components here
```

### Problem: Cluster labels overlap

**Solution**: Add padding and spacing
```python
graph_attr={
    "pad": "1.0",
    "nodesep": "1.0",
}
```

### Problem: Arrow labels overlapping edges

**Solution**: Use longer arrows with more rank separation
```python
graph_attr={"ranksep": "2.5"}
```

---

## Edge Labels & Styling

```python
from diagrams import Edge

# Basic label
node1 >> Edge(label="transforms") >> node2

# Styled edge
node1 >> Edge(
    label="critical path",
    color="red",
    style="bold",
    penwidth="2.0"
) >> node2

# Dashed edge
node1 >> Edge(style="dashed") >> node2
```

---

## Cluster Styling

```python
with Cluster("Production",
             graph_attr={
                 "bgcolor": "#f0f0f0",     # Light gray background
                 "style": "rounded",        # Rounded corners
                 "pencolor": "#333333",     # Border color
                 "penwidth": "2.0",         # Border width
                 "fontsize": "16",          # Cluster label size
             }):
```

---

## Multiple Output Formats

```python
with Diagram("Title",
             filename="output",
             outformat=["png", "svg", "dot"]):  # Generate all three
```

---

## DPI and Resolution

For high-quality exports:

```python
graph_attr={
    "dpi": "300",  # Print quality
}
```

| DPI | Use Case |
|-----|----------|
| 72 | Screen viewing |
| 150 | Standard quality |
| 300 | Print quality |
| 600 | Publication quality |

---

## Debugging Layout

1. **Generate DOT file**: Set `outformat="dot"` to see Graphviz source
2. **Use Graphviz directly**: `dot -Tpng -Kneato diagram.dot -o output.png`
3. **Try different engines**: `neato`, `fdp`, `circo`, `twopi` for different layouts

```python
# Generate DOT for debugging
with Diagram("Title", filename="debug", outformat="dot"):
    ...
```
