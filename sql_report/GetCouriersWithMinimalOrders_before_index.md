| QUERY PLAN |
| :--- |
| Seq Scan on couriers  \(cost=1495.35..2995.35 rows=963 width=8\) \(actual time=6.859..10.335 rows=1800 loops=1\) |
|   Filter: \(\(status = 'available'::text\) AND \(order\_count = $0\)\) |
|   Rows Removed by Filter: 48200 |
|   Buffers: shared hit=1500 |
|   InitPlan 1 \(returns $0\) |
|     -&gt;  Aggregate  \(cost=1495.34..1495.35 rows=1 width=8\) \(actual time=6.850..6.851 rows=1 loops=1\) |
|           Buffers: shared hit=750 |
|           -&gt;  Seq Scan on couriers couriers\_1  \(cost=0.00..1375.00 rows=48135 width=8\) \(actual time=0.001..5.111 rows=48100 loops=1\) |
|                 Filter: \(status = 'available'::text\) |
|                 Rows Removed by Filter: 1900 |
|                 Buffers: shared hit=750 |
| Planning Time: 0.084 ms |
| Execution Time: 10.425 ms |
