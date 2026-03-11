# Chart Interpretation Guide for Databricks AI/BI Dashboards

This guide helps you analyze and interpret visualizations from Databricks AI/BI dashboard screenshots.

## Table of Contents

- [Chart Types](#chart-types): Counter, Bar, Line, Area, Pie, Scatter, Histogram, Table
- [Visual Encodings](#visual-encodings): Color, Size, Position
- [Legend](#legend): Decoding chart legends
- [Axes](#axes): X-Axis and Y-Axis interpretation
- [Trend Analysis](#trend-analysis): Identifying and describing trends
- [Comparison Patterns](#comparison-patterns): Relative, temporal, segment comparisons
- [Identifying Insights](#identifying-insights): Key questions and insight types
- [Common Mistakes to Avoid](#common-mistakes-to-avoid)
- [Practical Examples](#practical-examples): Bar chart, line chart, counter analysis
- [Tips for Screenshot Analysis](#tips-for-screenshot-analysis)

## Chart Types

### Counter (KPI Widget)

**Appearance**: Large number, often with label and optional sparkline

**What to look for**:
- **Main value**: The primary metric (e.g., "$1.2M", "543 users")
- **Label**: What the metric represents
- **Trend indicator**: Arrow up/down or percentage change
- **Comparison**: "vs last month" or similar context
- **Color**: Green (positive), red (negative), neutral

**Example analysis**:
- "Total Revenue counter shows $1.2M, up 15% from last month (green indicator)"
- "Active Users: 543, down 8% (red arrow)"

### Bar Chart

**Appearance**: Vertical or horizontal bars representing categories

**What to look for**:
- **X-axis**: Categories (products, regions, time periods)
- **Y-axis**: Values (sales, counts, percentages)
- **Bar colors**: Single color or color-coded by category
- **Bar heights**: Compare relative magnitudes
- **Labels**: Data labels on/above bars
- **Sorted**: Are bars sorted by value or alphabetically?

**Analysis patterns**:
- **Comparison**: "Product A leads with 450 units, followed by Product B (320) and Product C (180)"
- **Distribution**: "Sales are heavily concentrated in the top 3 categories (80% of total)"
- **Outliers**: "One category significantly outperforms others"

### Line Chart

**Appearance**: Connected points showing trends over time

**What to look for**:
- **X-axis**: Time dimension (dates, months, quarters)
- **Y-axis**: Metric values
- **Multiple lines**: Different series (products, regions)
- **Line colors**: Each series has distinct color
- **Trend direction**: Upward, downward, flat, cyclical
- **Intersections**: Where lines cross (market share changes)
- **Gaps**: Missing data points

**Analysis patterns**:
- **Trend**: "Revenue shows consistent upward trend from Jan to Dec, growing ~5% month-over-month"
- **Seasonality**: "Sales peak in Q4 every year, with summer slump in July-Aug"
- **Volatility**: "Metric is highly variable with large swings week-to-week"
- **Crossover**: "Product A overtook Product B in market share in June"

### Area Chart

**Appearance**: Like line chart but area under line is filled

**What to look for**:
- **Stacked vs unstacked**: Stacked shows cumulative, unstacked shows overlapping
- **Total**: Top edge of stacked area shows overall sum
- **Proportions**: Relative thickness of each area shows contribution
- **Colors**: Each series has distinct fill color

**Analysis patterns**:
- **Composition**: "Category A contributes 40-50% of total throughout the period"
- **Growth**: "Total (top line) has grown 3x from Q1 to Q4"
- **Shifts**: "Category B's share decreased while Category C's increased"

### Pie Chart

**Appearance**: Circle divided into slices

**What to look for**:
- **Slice sizes**: Proportional to values
- **Colors**: Each category has distinct color
- **Labels**: Category names and percentages
- **Dominant slice**: Largest segment
- **Fragmentation**: Many small slices or few large ones

**Analysis patterns**:
- **Majority**: "Category A represents 65% of total, significantly larger than others"
- **Distribution**: "Evenly split between 4 categories (20-28% each)"
- **Long tail**: "Top 3 categories account for 80%, remaining 10 categories split 20%"

### Scatter Plot

**Appearance**: Points plotted on X-Y axes

**What to look for**:
- **X and Y axes**: What dimensions are being compared
- **Point colors**: Additional category dimension
- **Point sizes**: Optional fourth dimension (bubble chart)
- **Patterns**: Clusters, linear relationships, outliers
- **Correlation**: Positive, negative, or no relationship
- **Outliers**: Points far from main cluster

**Analysis patterns**:
- **Correlation**: "Strong positive correlation between advertising spend and sales (R² ~0.85)"
- **Clustering**: "Two distinct clusters: high-price/low-volume and low-price/high-volume"
- **Outliers**: "One product has unusually high margin despite low volume"

### Histogram

**Appearance**: Bars representing frequency distribution

**What to look for**:
- **X-axis**: Bins/ranges (e.g., age groups, price ranges)
- **Y-axis**: Frequency/count in each bin
- **Shape**: Normal, skewed, bimodal, uniform
- **Peak**: Most common range
- **Spread**: How widely distributed

**Analysis patterns**:
- **Normal distribution**: "Customer ages follow bell curve, centered around 35-44"
- **Right-skewed**: "Most transactions are $10-50, with long tail up to $500"
- **Bimodal**: "Two peaks suggest two distinct customer segments"

### Table / Data Grid

**Appearance**: Rows and columns of data

**What to look for**:
- **Column headers**: What data is shown
- **Sort order**: How is data sorted (by which column)
- **Conditional formatting**: Color-coded cells (red/green, heat maps)
- **Totals/aggregates**: Summary rows
- **Pagination**: Are all rows shown or paginated?
- **Number formatting**: Currency, percentages, decimals

**Analysis patterns**:
- **Top performers**: "Top 5 sales reps by revenue: Alice ($150K), Bob ($145K)..."
- **Patterns**: "All top 10 items are in Electronics category"
- **Outliers**: "One transaction is 10x larger than typical"

## Visual Encodings

### Color

**Purpose**: Distinguish categories, show thresholds, indicate status

**Common patterns**:
- **Categorical**: Different colors for different series/categories
- **Sequential**: Light to dark for low to high values
- **Diverging**: Two colors for positive/negative or above/below average
- **Status**: Green (good), yellow (warning), red (bad)
- **Brand colors**: Databricks orange, blue, teal

**Interpretation**:
- "Green bars indicate targets met, red bars show targets missed"
- "Darker blue represents higher values (heat map encoding)"
- "Each product line has consistent color across all charts"

### Size

**Purpose**: Encode magnitude or importance

**Common patterns**:
- **Bar length**: Proportional to value
- **Circle size**: In bubble charts, represents third dimension
- **Font size**: Emphasize important metrics

**Interpretation**:
- "Larger circles represent higher revenue products"
- "Bar heights directly compare sales volumes"

### Position

**Purpose**: Show relationships, comparisons, spatial data

**Common patterns**:
- **X/Y axes**: Plot two dimensions against each other
- **Stacking**: Show cumulative totals
- **Grouping**: Related items positioned together

**Interpretation**:
- "Higher position indicates better performance"
- "Closer proximity suggests correlation"

## Legend

**Purpose**: Decode color/size/shape meanings

**What to look for**:
- **Position**: Usually top-right, right side, or bottom
- **Series names**: What each color represents
- **Order**: May indicate relative importance
- **Interactive**: Sometimes clickable to filter

**Interpretation**:
- "Legend shows 3 regions: North America (blue), Europe (orange), Asia (green)"
- "Series ordered by total contribution (descending)"

## Axes

### X-Axis (Horizontal)

**Common encodings**:
- **Categorical**: Products, regions, names (bar charts)
- **Temporal**: Dates, months, quarters (line/area charts)
- **Quantitative**: Price ranges, age groups (histograms, scatter)

**What to check**:
- **Labels**: Are all categories shown or truncated?
- **Rotation**: Angled labels for readability
- **Scale**: Linear, logarithmic, time-based
- **Range**: Start and end values

### Y-Axis (Vertical)

**Common encodings**:
- **Quantitative**: Sales, counts, percentages, metrics
- **Zero baseline**: Does axis start at zero or different value?

**What to check**:
- **Scale**: Linear vs logarithmic
- **Range**: Min and max values
- **Zero**: Starting point affects visual perception
- **Units**: Currency, percentages, raw numbers
- **Dual axes**: Some charts have two Y-axes (left and right)

**Important**: Non-zero baselines can exaggerate differences!

## Trend Analysis

### Trend Types

1. **Upward**: Consistent increase over time
   - "Growing 10-15% quarter-over-quarter"
   - "Accelerating growth rate"

2. **Downward**: Consistent decrease
   - "Declining 5% per month"
   - "Steep drop followed by stabilization"

3. **Flat/Stable**: Little change
   - "Holding steady around $500K monthly"
   - "Slight variations but no clear direction"

4. **Cyclical/Seasonal**: Regular patterns
   - "Peaks in December (holiday season), dips in February"
   - "Weekly pattern: higher on weekends"

5. **Volatile**: Large fluctuations
   - "Highly variable, ranging from $100K to $400K"
   - "Unstable with no clear pattern"

### Growth Calculation

From visual inspection:
- **Approximate percentage change**: Compare start and end heights
- **Compound growth**: If line curves upward, growth is accelerating
- **Rate**: Steep line = fast change, gradual line = slow change

**Example**: "Line starts at 100 and ends at 150 over 12 months = ~50% annual growth, or ~4% monthly"

## Comparison Patterns

### Relative Comparison

- **Ranking**: "A is largest, then B, then C"
- **Magnitude**: "A is 3x larger than B"
- **Percentage**: "A represents 45% of total, B is 30%, C is 25%"
- **Gap**: "Significant gap between top performer and rest"

### Temporal Comparison

- **Year-over-year**: "2024 vs 2023"
- **Month-over-month**: "December vs November"
- **Period comparison**: "Q4 vs Q3"

### Segment Comparison

- **Geographic**: "North America outperforms other regions"
- **Demographic**: "25-34 age group is most active"
- **Product**: "Premium tier has higher engagement"

## Identifying Insights

### Key Questions

1. **What's the main message?**
   - Is one category dominant?
   - Is there a clear trend?
   - Are there unexpected patterns?

2. **What's changing?**
   - Which direction?
   - How fast?
   - Since when?

3. **What's notable?**
   - Outliers?
   - Anomalies?
   - Records (highs/lows)?

4. **What's the context?**
   - Are values good or bad?
   - How does this compare to targets?
   - What's driving the pattern?

### Insight Types

**Performance insights**:
- "Revenue exceeded target by 12%"
- "Conversion rate improved from 2.1% to 2.8%"

**Trend insights**:
- "User growth has accelerated in last 3 months"
- "Churn rate declining steadily since Q2"

**Comparison insights**:
- "Product A outselling Product B 2:1"
- "West region growing faster than East (15% vs 8%)"

**Anomaly insights**:
- "Unusual spike in traffic on Nov 15 (2x normal)"
- "Sales dropped 40% in one week (investigate)"

**Composition insights**:
- "Mobile traffic now exceeds desktop (52% vs 48%)"
- "Three products account for 75% of revenue"

## Common Mistakes to Avoid

1. **Ignoring scale**: Non-zero baseline can exaggerate differences
2. **Confusing correlation and causation**: Just because lines move together doesn't mean one causes the other
3. **Over-interpreting noise**: Random variation vs real signals
4. **Missing context**: Numbers need benchmarks (targets, historical, competitors)
5. **Ignoring sample size**: Small samples can be misleading
6. **Not checking units**: Percentages vs absolute numbers, currency vs counts

## Practical Examples

### Example 1: Bar Chart Analysis

**Visual**: Horizontal bar chart with 10 bars of varying lengths, colored by category

**Analysis**:
```
Chart Type: Horizontal bar chart comparing sales by product category
Encoding: X-axis = sales ($), Y-axis = categories, color = category type
Key Findings:
- Electronics leads with $450K (32% of total)
- Top 3 categories (Electronics, Clothing, Home) represent 65% of sales
- Bottom 5 categories combined < 15% of total (long tail)
- Clear hierarchy with exponential drop-off after top 3
Insight: Focus inventory and marketing on top 3 performing categories
```

### Example 2: Line Chart Analysis

**Visual**: Line chart with 3 colored lines over 12 months

**Analysis**:
```
Chart Type: Multi-series line chart showing revenue trends by region
Encoding: X = months (Jan-Dec), Y = revenue ($), color = region
Key Findings:
- All regions show upward trends (positive growth)
- North America (blue): Steady growth, 25% increase year-over-year
- Europe (orange): Accelerating growth, especially last 3 months (40% YoY)
- Asia (green): Slower growth, relatively flat (8% YoY)
- Europe overtook Asia in June, now approaching North America
Seasonality: All regions show Q4 spike (holiday season)
Insight: Europe is star performer; investigate Asia's slow growth
```

### Example 3: Counter + Sparkline

**Visual**: Large "$1.2M" with small line chart below, green up-arrow, "+15%"

**Analysis**:
```
Widget Type: KPI counter with trend indicator
Primary Metric: $1.2M total revenue
Comparison: +15% vs previous period (green indicates positive)
Trend: Sparkline shows consistent upward trajectory over time
Context: Strong performance, exceeding baseline by significant margin
Status: Healthy growth, on positive trajectory
```

## Tips for Screenshot Analysis

1. **Read title and labels first**: Understand what you're looking at
2. **Check axes and legends**: Know what encodings mean
3. **Identify chart type**: Different types = different insights
4. **Look for patterns**: Trends, outliers, clusters, gaps
5. **Compare magnitudes**: Relative sizes, not just order
6. **Note colors and visual cues**: Status indicators, thresholds
7. **Consider context**: Time periods, filters applied, comparisons
8. **Extract numbers when possible**: For precision
9. **Synthesize**: Combine multiple charts for complete picture
10. **State confidence**: Note if visual analysis is approximate

