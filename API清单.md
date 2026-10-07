# API 清单

基址：`http://host:8080`。业务接口只使用 GET 和 POST，每个操作一条独立 path。

## 通用返回

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

| code | HTTP | 含义 |
| --- | --- | --- |
| 0 | 200 | 成功 |
| 400 | 400 | 参数错误 |
| 401 | 401 | 未登录、令牌无效，或用户已不存在 |
| 404 | 404 | 记录不存在。域名能解析出 brand，但 subpad 不存在时也是 404 |
| 409 | 409 | 唯一约束冲突 |
| 500 | 500 | 服务器错误 |
| 503 | 503 | 未配置 MySQL |

失败示例：

```json
{
  "code": 404,
  "message": "not found",
  "data": null
}
```

删除成功时 `data` 为 `null`。列表的 `limit` 小于等于 0 时按 20 条，最大 200。修改接口会整行替换，未传的字段按零值写入；带创建时间的表会保留原来的 `created_at`。

除 `POST /api/user_info/login` 和 `POST /api/user_info/create` 外，`/api` 请求都要带登录令牌：

```
Authorization: Bearer <jwtToken>
```

`JwtFilter` 校验令牌后查 `user_info`，把当前用户放进请求上下文。令牌默认 72 小时有效。

`DomainFilter` 读取 Host。形如 `foods.launch.o1.local` 时，取出 `foods`，按 `subpad_info.brand` 查询并放进请求上下文。主机名不带这个后缀时跳过，例如 `localhost`。后缀由配置 `domain.suffix` 指定，默认 `launch.o1.local`。

## GET /health

请求：无。

响应：

```json
{
  "status": "ok",
  "mysql": true,
  "eth": false,
  "cron": true
}
```

## POST /api/user_info/login

不需要令牌。

请求：

```json
{
  "username": "alice",
  "password": "secret"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "jwtToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

用户名或密码错误时：

```json
{
  "code": 401,
  "message": "invalid username or password",
  "data": null
}
```

## POST /api/user_info/create

请求：

```json
{
  "username": "alice",
  "password": "secret",
  "fee_addr": "0xfee"
}
```

`username`、`password` 必填。密码明文传入，存储为 bcrypt。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "username": "alice",
    "fee_addr": "0xfee"
  }
}
```

响应不返回 `password`。

## POST /api/user_info/update

请求：

```json
{
  "id": 1,
  "username": "alice",
  "password": "",
  "fee_addr": "0xnew"
}
```

`password` 不传或传空则保持原密码。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "username": "alice",
    "fee_addr": "0xnew"
  }
}
```

## POST /api/user_info/delete

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## GET /api/user_info/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "username": "alice",
    "fee_addr": "0xfee"
  }
}
```

## GET /api/user_info/list

请求：查询参数 `username=alice&offset=0&limit=20`。`username` 可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "username": "alice",
      "fee_addr": "0xfee"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/token_info/create

请求：

```json
{
  "subpad_id": 7,
  "pool_id": "pool-1",
  "creator": "0xcreator",
  "chainid": 1,
  "token_addr": "0xtoken",
  "token_name": "Demo",
  "token_symbol": "AAA",
  "quote_token_addr": "0xusdc",
  "quote_token_symbol": "USDC",
  "launch_supply": 1000000,
  "tick_spacing": 60
}
```

`subpad_id` 不传则为空。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "subpad_id": 7,
    "pool_id": "pool-1",
    "creator": "0xcreator",
    "chainid": 1,
    "token_addr": "0xtoken",
    "token_name": "Demo",
    "token_symbol": "AAA",
    "quote_token_addr": "0xusdc",
    "quote_token_symbol": "USDC",
    "launch_supply": 1000000,
    "tick_spacing": 60
  }
}
```

## POST /api/token_info/update

请求：

```json
{
  "id": 1,
  "subpad_id": 7,
  "pool_id": "pool-1",
  "creator": "0xcreator",
  "chainid": 1,
  "token_addr": "0xtoken",
  "token_name": "Demo",
  "token_symbol": "AAA",
  "quote_token_addr": "0xusdc",
  "quote_token_symbol": "USDC",
  "launch_supply": 2000000,
  "tick_spacing": 60
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "subpad_id": 7,
    "pool_id": "pool-1",
    "creator": "0xcreator",
    "chainid": 1,
    "token_addr": "0xtoken",
    "token_name": "Demo",
    "token_symbol": "AAA",
    "quote_token_addr": "0xusdc",
    "quote_token_symbol": "USDC",
    "launch_supply": 2000000,
    "tick_spacing": 60
  }
}
```

## POST /api/token_info/delete

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## GET /api/token_info/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "subpad_id": 7,
    "pool_id": "pool-1",
    "creator": "0xcreator",
    "chainid": 1,
    "token_addr": "0xtoken",
    "token_name": "Demo",
    "token_symbol": "AAA",
    "quote_token_addr": "0xusdc",
    "quote_token_symbol": "USDC",
    "launch_supply": 1000000,
    "tick_spacing": 60
  }
}
```

## GET /api/token_info/list

请求：查询参数 `subpad_id=7&pool_id=pool-1&creator=0xcreator&chainid=1&token_addr=0xtoken&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "subpad_id": 7,
      "pool_id": "pool-1",
      "creator": "0xcreator",
      "chainid": 1,
      "token_addr": "0xtoken",
      "token_name": "Demo",
      "token_symbol": "AAA",
      "quote_token_addr": "0xusdc",
      "quote_token_symbol": "USDC",
      "launch_supply": 1000000,
      "tick_spacing": 60
    }
  ]
}
```

按 `id` 倒序。

## POST /api/subpad_info/create

请求：

```json
{
  "user_id": 3,
  "user_addr": "0xuser",
  "brand": "foods",
  "name_full": "Foods Pad",
  "status": 1,
  "swap_type": "mockSwap",
  "description": "详细描述"
}
```

`swap_type`：`mockSwap` 模拟，`uniSwap` 真实。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "user_id": 3,
    "user_addr": "0xuser",
    "brand": "foods",
    "name_full": "Foods Pad",
    "status": 1,
    "swap_type": "mockSwap",
    "description": "详细描述",
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/subpad_info/update

请求：

```json
{
  "id": 1,
  "user_id": 3,
  "user_addr": "0xuser",
  "brand": "foods",
  "name_full": "Foods Pad",
  "status": 1,
  "swap_type": "uniSwap",
  "description": "更新后的描述"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "user_id": 3,
    "user_addr": "0xuser",
    "brand": "foods",
    "name_full": "Foods Pad",
    "status": 1,
    "swap_type": "uniSwap",
    "description": "更新后的描述",
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T16:00:00+08:00"
  }
}
```

`created_at` 保持创建时的值。

## POST /api/subpad_info/delete

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## GET /api/subpad_info/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "user_id": 3,
    "user_addr": "0xuser",
    "brand": "foods",
    "name_full": "Foods Pad",
    "status": 1,
    "swap_type": "mockSwap",
    "description": "详细描述",
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/subpad_info/list

请求：查询参数 `user_id=3&brand=foods&status=1&swap_type=mockSwap&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "user_id": 3,
      "user_addr": "0xuser",
      "brand": "foods",
      "name_full": "Foods Pad",
      "status": 1,
      "swap_type": "mockSwap",
      "description": "详细描述",
      "created_at": "2026-10-07T15:09:00+08:00",
      "updated_at": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/fee_info/create

请求：

```json
{
  "chainid": 1,
  "pool_id": "pool-1",
  "tx_hash": "0xtx",
  "fee_type": "platform",
  "fee_token": "0xusdc",
  "fee_decimal": 6,
  "fee_amount": 100,
  "fee_to": "0xfee"
}
```

`fee_type`：`platform` 平台费，`tokencreator` 创建者费，`subpad` 子 pad 费。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "pool_id": "pool-1",
    "tx_hash": "0xtx",
    "fee_type": "platform",
    "fee_token": "0xusdc",
    "fee_decimal": 6,
    "fee_amount": 100,
    "fee_to": "0xfee",
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/fee_info/update

请求：

```json
{
  "id": 1,
  "chainid": 1,
  "pool_id": "pool-1",
  "tx_hash": "0xtx",
  "fee_type": "subpad",
  "fee_token": "0xusdc",
  "fee_decimal": 6,
  "fee_amount": 200,
  "fee_to": "0xfee"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "pool_id": "pool-1",
    "tx_hash": "0xtx",
    "fee_type": "subpad",
    "fee_token": "0xusdc",
    "fee_decimal": 6,
    "fee_amount": 200,
    "fee_to": "0xfee",
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T16:00:00+08:00"
  }
}
```

## POST /api/fee_info/delete

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## GET /api/fee_info/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "pool_id": "pool-1",
    "tx_hash": "0xtx",
    "fee_type": "platform",
    "fee_token": "0xusdc",
    "fee_decimal": 6,
    "fee_amount": 100,
    "fee_to": "0xfee",
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/fee_info/list

请求：查询参数 `chainid=1&pool_id=pool-1&tx_hash=0xtx&fee_type=platform&fee_to=0xfee&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainid": 1,
      "pool_id": "pool-1",
      "tx_hash": "0xtx",
      "fee_type": "platform",
      "fee_token": "0xusdc",
      "fee_decimal": 6,
      "fee_amount": 100,
      "fee_to": "0xfee",
      "created_at": "2026-10-07T15:09:00+08:00",
      "updated_at": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/sync_event/create

请求：

```json
{
  "chainid": 1,
  "block_number": 10,
  "block_hash": "0xblock",
  "tx_hash": "0xhash",
  "tx_index": 0,
  "log_index": 2,
  "contract_addr": "0xcontract",
  "topics": "0xtopic",
  "data": "0xdata",
  "removed": 0
}
```

`removed`：0 未移除，1 因链重组已移除。`chainid + tx_hash + log_index` 唯一，重复写入返回 409。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "block_number": 10,
    "block_hash": "0xblock",
    "tx_hash": "0xhash",
    "tx_index": 0,
    "log_index": 2,
    "contract_addr": "0xcontract",
    "topics": "0xtopic",
    "data": "0xdata",
    "removed": 0,
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/sync_event/update

请求：

```json
{
  "id": 1,
  "chainid": 1,
  "block_number": 10,
  "block_hash": "0xblock",
  "tx_hash": "0xhash",
  "tx_index": 0,
  "log_index": 2,
  "contract_addr": "0xcontract",
  "topics": "0xtopic",
  "data": "0xdata",
  "removed": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "block_number": 10,
    "block_hash": "0xblock",
    "tx_hash": "0xhash",
    "tx_index": 0,
    "log_index": 2,
    "contract_addr": "0xcontract",
    "topics": "0xtopic",
    "data": "0xdata",
    "removed": 1,
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T16:00:00+08:00"
  }
}
```

## POST /api/sync_event/delete

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## GET /api/sync_event/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "block_number": 10,
    "block_hash": "0xblock",
    "tx_hash": "0xhash",
    "tx_index": 0,
    "log_index": 2,
    "contract_addr": "0xcontract",
    "topics": "0xtopic",
    "data": "0xdata",
    "removed": 0,
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/sync_event/list

请求：查询参数 `chainid=1&tx_hash=0xhash&contract_addr=0xcontract&removed=0&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainid": 1,
      "block_number": 10,
      "block_hash": "0xblock",
      "tx_hash": "0xhash",
      "tx_index": 0,
      "log_index": 2,
      "contract_addr": "0xcontract",
      "topics": "0xtopic",
      "data": "0xdata",
      "removed": 0,
      "created_at": "2026-10-07T15:09:00+08:00",
      "updated_at": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `block_number`、`log_index` 倒序。

## POST /api/sync_cursor/create

请求：

```json
{
  "chainid": 1,
  "block_number": 100
}
```

`chainid` 唯一，同一条链重复创建返回 409。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "block_number": 100,
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/sync_cursor/update

请求：

```json
{
  "id": 1,
  "chainid": 1,
  "block_number": 250
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "block_number": 250,
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T16:00:00+08:00"
  }
}
```

## POST /api/sync_cursor/delete

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## GET /api/sync_cursor/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainid": 1,
    "block_number": 100,
    "created_at": "2026-10-07T15:09:00+08:00",
    "updated_at": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/sync_cursor/list

请求：查询参数 `chainid=1&offset=0&limit=20`。`chainid` 可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainid": 1,
      "block_number": 250,
      "created_at": "2026-10-07T15:09:00+08:00",
      "updated_at": "2026-10-07T16:00:00+08:00"
    }
  ]
}
```

按 `chainid` 升序。
