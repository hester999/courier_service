| QUERY PLAN |
| :--- |
| Index Scan using couriers\_status\_idx on couriers  \(cost=0.29..551.31 rows=1885 width=67\) \(actual time=0.016..0.433 rows=1900 loops=1\) |
|   Index Cond: \(status = 'busy'::text\) |
|   Buffers: shared hit=189 |
| Planning: |
|   Buffers: shared hit=29 |
| Planning Time: 0.145 ms |
| Execution Time: 0.488 ms |
