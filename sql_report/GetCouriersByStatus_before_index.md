| QUERY PLAN |
| :--- |
| Index Scan using couriers\_status\_idx on couriers  \(cost=0.29..665.43 rows=2398 width=67\) \(actual time=1.139..19.878 rows=1900 loops=1\) |
|   Index Cond: \(status = 'busy'::text\) |
|   Buffers: shared hit=3 read=186 |
| Planning: |
|   Buffers: shared hit=77 read=5 dirtied=2 |
| Planning Time: 1.728 ms |
| Execution Time: 19.965 ms |
