| QUERY PLAN |
| :--- |
| Bitmap Heap Scan on couriers  \(cost=14.50..818.96 rows=961 width=8\) \(actual time=0.220..0.906 rows=1800 loops=1\) |
|   Recheck Cond: \(\(status = 'available'::text\) AND \(order\_count = $1\)\) |
|   Heap Blocks: exact=749 |
|   Buffers: shared hit=753 read=3 |
|   InitPlan 2 \(returns $1\) |
|     -&gt;  Result  \(cost=0.35..0.36 rows=1 width=8\) \(actual time=0.039..0.040 rows=1 loops=1\) |
|           Buffers: shared hit=2 read=2 |
|           InitPlan 1 \(returns $0\) |
|             -&gt;  Limit  \(cost=0.29..0.35 rows=1 width=8\) \(actual time=0.037..0.038 rows=1 loops=1\) |
|                   Buffers: shared hit=2 read=2 |
|                   -&gt;  Index Only Scan using couriers\_available\_order\_count\_idx on couriers couriers\_1  \(cost=0.29..2818.78 rows=48047 width=8\) \(actual time=0.037..0.037 rows=1 loops=1\) |
|                         Index Cond: \(\(status = 'available'::text\) AND \(order\_count IS NOT NULL\)\) |
|                         Heap Fetches: 1 |
|                         Buffers: shared hit=2 read=2 |
|   -&gt;  Bitmap Index Scan on couriers\_available\_order\_count\_idx  \(cost=0.00..13.90 rows=961 width=0\) \(actual time=0.146..0.146 rows=1800 loops=1\) |
|         Index Cond: \(\(status = 'available'::text\) AND \(order\_count = $1\)\) |
|         Buffers: shared hit=4 read=3 |
| Planning: |
|   Buffers: shared hit=14 read=1 |
| Planning Time: 0.273 ms |
| Execution Time: 0.984 ms |
