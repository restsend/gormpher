# Gormpher Help Docs

## WebObject Default Handlers

- Get
- Query
- Create
- Edit
- Delete
- Batch Delete

## Actions

- Edit
- Filter
- Search
- Order

## Query Object

QueryForm:

| Name  |  Type |  Desc |  Default  |
|---|---| --- |  --- |
|  pos | number  |    |  0  |
|  limit | number  |    |  50  |
|  keyword | number  |    | ""  |
|  filters | <a class="link">[]Filter</a>  |   | null  |
|  orders | <a class="link">[]Order |    | null |

Filter:

| Name  |  Op |  Desc |
|---|---| --- |
|  name | string |    |
|  op | string |  `=, <>, int, not_in, >, >=, <, <=`  |
|  value | string|    |

Order:

| Name  |  Op |  Desc |
|---|---| --- |
|  name | string |    |
|  op | string |  `asc, desc` |

## Query Result

QueryResult:

| Name  |  Type |  Desc |
|---|---| --- |
|  pos | number  |    |
|  limit | number  |    |
|  keyword | string |    |
|  total | number  |    |
|  items | []object |  golang struct model |
