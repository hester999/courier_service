| QUERY PLAN |
| :--- |
| HashAggregate  \(cost=5848.28..5884.85 rows=813 width=4\) \(actual time=46.092..46.095 rows=0 loops=1\) |
|   Group Key: c.id |
|   Filter: \(max\(d.deadline\) &lt; now\(\)\) |
|   Batches: 1  Memory Usage: 369kB |
|   Rows Removed by Filter: 2500 |
|   Buffers: shared hit=2619 |
|   -&gt;  Hash Join  \(cost=1404.47..5799.52 rows=9752 width=12\) \(actual time=5.486..43.357 rows=10000 loops=1\) |
|         Hash Cond: \(d.courier\_id = c.id\) |
|         Buffers: shared hit=2619 |
|         -&gt;  Seq Scan on delivery d  \(cost=0.00..3870.00 rows=200000 width=16\) \(actual time=0.005..14.132 rows=200000 loops=1\) |
|               Buffers: shared hit=1870 |
|         -&gt;  Hash  \(cost=1374.00..1374.00 rows=2438 width=4\) \(actual time=5.243..5.244 rows=2500 loops=1\) |
|               Buckets: 4096  Batches: 1  Memory Usage: 120kB |
|               Buffers: shared hit=749 |
|               -&gt;  Seq Scan on couriers c  \(cost=0.00..1374.00 rows=2438 width=4\) \(actual time=0.137..4.924 rows=2500 loops=1\) |
|                     Filter: \(status = 'busy'::text\) |
|                     Rows Removed by Filter: 47500 |
|                     Buffers: shared hit=749 |
| Planning: |
|   Buffers: shared hit=189 |
| Planning Time: 0.456 ms |
| Execution Time: 46.181 ms |
