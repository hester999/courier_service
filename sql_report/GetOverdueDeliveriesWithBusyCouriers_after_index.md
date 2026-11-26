| QUERY PLAN |
| :--- |
| HashAggregate  \(cost=3417.86..3453.78 rows=798 width=4\) \(actual time=6.986..6.986 rows=0 loops=1\) |
|   Group Key: c.id |
|   Filter: \(max\(d.deadline\) &lt; now\(\)\) |
|   Batches: 1  Memory Usage: 369kB |
|   Rows Removed by Filter: 2425 |
|   Buffers: shared hit=7468 |
|   -&gt;  Nested Loop  \(cost=0.71..3369.96 rows=9580 width=12\) \(actual time=0.035..5.220 rows=9700 loops=1\) |
|         Buffers: shared hit=7468 |
|         -&gt;  Index Scan using couriers\_status\_idx on couriers c  \(cost=0.29..664.61 rows=2395 width=4\) \(actual time=0.011..0.707 rows=2425 loops=1\) |
|               Index Cond: \(status = 'busy'::text\) |
|               Buffers: shared hit=189 |
|         -&gt;  Index Only Scan using delivery\_courier\_deadline\_idx on delivery d  \(cost=0.42..1.09 rows=4 width=16\) \(actual time=0.001..0.001 rows=4 loops=2425\) |
|               Index Cond: \(courier\_id = c.id\) |
|               Heap Fetches: 0 |
|               Buffers: shared hit=7279 |
| Planning: |
|   Buffers: shared hit=48 |
| Planning Time: 0.314 ms |
| Execution Time: 7.016 ms |
