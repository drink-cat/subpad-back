# API 清单

基址：`http://host:8080`。业务接口只使用 GET 和 POST，每个操作一条独立 path。请求 JSON、响应 JSON 和查询参数都使用驼峰，例如 `feeAddr`、`createdAt`、`chainId`。数据库列名仍使用下划线。

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

删除成功时 `data` 为 `null`。列表的 `limit` 小于等于 0 时按 20 条，最大 200。修改接口会整行替换，未传的字段按零值写入；带创建时间的表会保留原来的 `createdAt`。

除 `POST /api/user_info/login` 和 `POST /api/user_info/create` 外，`/api` 请求都要带登录令牌：

```
Authorization: Bearer <jwtToken>
```

`JwtFilter` 校验令牌后查 `user_info`，把当前用户放进请求上下文。令牌默认 72 小时有效。

`DomainFilter` 读取主机名。请求带 `X-Forwarded-Host` 时用它（前端把接口代理到本机时，原始域名在这个头里），否则用 `Host`。形如 `foods.launch.o1.local` 时，取出 `foods`，按 `subpad_info.brand` 查询并放进请求上下文。查到后把这一行 JSON 放进响应头 `X-Subpad-Info`，字段与 `subpad_info` 接口相同。主机名不带这个后缀时跳过，不写这个头，例如 `localhost`。后缀由配置 `domain.suffix` 指定，默认 `launch.o1.local`。

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

## GET /api/config

不需要令牌。给前端读取链和合约配置。不返回扫块起始块、确认数、定时周期，也不返回数据库和登录密钥。

请求：无。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "quoteToken": {
      "localUsdc": "0xcB0f2a13098f8e841e6Adfa5B17Ec00508b27665",
      "sepoliaUsdc": "0x5728d6521217108001057f09271311987e81d5a0"
    },
    "syncLog": [
      {
        "name": "本地网",
        "chainId": 31337,
        "rpcUrl": "http://127.0.0.1:8545",
        "launchContract": "0x88D1aF96098a928eE278f162c1a84f339652f95b"
      }
    ]
  }
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
  "feeAddr": "0xfee"
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
    "feeAddr": "0xfee"
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
  "feeAddr": "0xnew"
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
    "feeAddr": "0xnew"
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
    "feeAddr": "0xfee"
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
      "feeAddr": "0xfee"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/token_info/create

发币时先写入一行。`userId` 由后端按当前登录用户写入，请求体里的 `userId` 不采用。此时 `tokenAddr` 等链上字段可以先空着。

请求：

```json
{
  "subpadId": 7,
  "creator": "0xcreator",
  "chainId": 1,
  "tokenName": "Demo",
  "tokenSymbol": "AAA",
  "quoteTokenAddr": "0xusdc",
  "quoteTokenSymbol": "USDC",
  "launchSupply": 1000000,
  "tickSpacing": 60
}
```

`subpadId` 不传则为空。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "userId": 1,
    "subpadId": 7,
    "poolId": "",
    "creator": "0xcreator",
    "chainId": 1,
    "tokenAddr": "",
    "tokenName": "Demo",
    "tokenSymbol": "AAA",
    "quoteTokenAddr": "0xusdc",
    "quoteTokenSymbol": "USDC",
    "launchSupply": 1000000,
    "tickSpacing": 60
  }
}
```

## POST /api/token_info/update

合约日志到达后，用这条接口补上 `tokenAddr`、`poolId` 等字段。`userId` 保持创建时的值，请求体里的 `userId` 不采用。

请求：

```json
{
  "id": 1,
  "subpadId": 7,
  "poolId": "pool-1",
  "creator": "0xcreator",
  "chainId": 1,
  "tokenAddr": "0xtoken",
  "tokenName": "Demo",
  "tokenSymbol": "AAA",
  "quoteTokenAddr": "0xusdc",
  "quoteTokenSymbol": "USDC",
  "launchSupply": 1000000,
  "tickSpacing": 60
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "userId": 1,
    "subpadId": 7,
    "poolId": "pool-1",
    "creator": "0xcreator",
    "chainId": 1,
    "tokenAddr": "0xtoken",
    "tokenName": "Demo",
    "tokenSymbol": "AAA",
    "quoteTokenAddr": "0xusdc",
    "quoteTokenSymbol": "USDC",
    "launchSupply": 1000000,
    "tickSpacing": 60
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
    "userId": 1,
    "subpadId": 7,
    "poolId": "pool-1",
    "creator": "0xcreator",
    "chainId": 1,
    "tokenAddr": "0xtoken",
    "tokenName": "Demo",
    "tokenSymbol": "AAA",
    "quoteTokenAddr": "0xusdc",
    "quoteTokenSymbol": "USDC",
    "launchSupply": 1000000,
    "tickSpacing": 60
  }
}
```

## GET /api/token_info/list

请求：查询参数 `userId=1&subpadId=7&poolId=pool-1&creator=0xcreator&chainId=1&tokenAddr=0xtoken&tokenSymbol=AAA&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "userId": 1,
      "subpadId": 7,
      "poolId": "pool-1",
      "creator": "0xcreator",
      "chainId": 1,
      "tokenAddr": "0xtoken",
      "tokenName": "Demo",
      "tokenSymbol": "AAA",
      "quoteTokenAddr": "0xusdc",
      "quoteTokenSymbol": "USDC",
      "launchSupply": 1000000,
      "tickSpacing": 60
    }
  ]
}
```

按 `id` 倒序。

## POST /api/subpad_info/create

请求：

```json
{
  "userId": 3,
  "brand": "foods",
  "nameFull": "Foods Pad",
  "status": 1,
  "swapType": "mockSwap",
  "description": "详细描述"
}
```

`swapType`：`mockSwap` 模拟，`uniSwap` 真实。请求里的 `feeAddr` 会被忽略。创建时后端写入当前登录用户的 `feeAddr`。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "userId": 3,
    "feeAddr": "0xfee",
    "brand": "foods",
    "nameFull": "Foods Pad",
    "status": 1,
    "swapType": "mockSwap",
    "description": "详细描述",
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/subpad_info/update

请求：

```json
{
  "id": 1,
  "userId": 3,
  "brand": "foods",
  "nameFull": "Foods Pad",
  "status": 1,
  "swapType": "uniSwap",
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
    "userId": 3,
    "feeAddr": "0xfee",
    "brand": "foods",
    "nameFull": "Foods Pad",
    "status": 1,
    "swapType": "uniSwap",
    "description": "更新后的描述",
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T16:00:00+08:00"
  }
}
```

`createdAt` 和 `feeAddr` 保持创建时的值。

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
    "userId": 3,
    "feeAddr": "0xfee",
    "brand": "foods",
    "nameFull": "Foods Pad",
    "status": 1,
    "swapType": "mockSwap",
    "description": "详细描述",
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/subpad_info/list

请求：查询参数 `userId=3&brand=foods&status=1&swapType=mockSwap&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "userId": 3,
      "feeAddr": "0xfee",
      "brand": "foods",
      "nameFull": "Foods Pad",
      "status": 1,
      "swapType": "mockSwap",
      "description": "详细描述",
      "createdAt": "2026-10-07T15:09:00+08:00",
      "updatedAt": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/fee_info/create

请求：

```json
{
  "chainId": 1,
  "poolId": "pool-1",
  "txHash": "0xtx",
  "feeType": "platform",
  "feeToken": "0xusdc",
  "feeDecimal": 6,
  "feeAmount": 100,
  "feeTo": "0xfee"
}
```

`feeType`：`platform` 平台费，`tokencreator` 创建者费，`subpad` 子 pad 费。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "poolId": "pool-1",
    "txHash": "0xtx",
    "feeType": "platform",
    "feeToken": "0xusdc",
    "feeDecimal": 6,
    "feeAmount": 100,
    "feeTo": "0xfee",
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/fee_info/update

请求：

```json
{
  "id": 1,
  "chainId": 1,
  "poolId": "pool-1",
  "txHash": "0xtx",
  "feeType": "subpad",
  "feeToken": "0xusdc",
  "feeDecimal": 6,
  "feeAmount": 200,
  "feeTo": "0xfee"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "poolId": "pool-1",
    "txHash": "0xtx",
    "feeType": "subpad",
    "feeToken": "0xusdc",
    "feeDecimal": 6,
    "feeAmount": 200,
    "feeTo": "0xfee",
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T16:00:00+08:00"
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
    "chainId": 1,
    "poolId": "pool-1",
    "txHash": "0xtx",
    "feeType": "platform",
    "feeToken": "0xusdc",
    "feeDecimal": 6,
    "feeAmount": 100,
    "feeTo": "0xfee",
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/fee_info/list

请求：查询参数 `chainId=1&poolId=pool-1&txHash=0xtx&feeType=platform&feeTo=0xfee&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainId": 1,
      "poolId": "pool-1",
      "txHash": "0xtx",
      "feeType": "platform",
      "feeToken": "0xusdc",
      "feeDecimal": 6,
      "feeAmount": 100,
      "feeTo": "0xfee",
      "createdAt": "2026-10-07T15:09:00+08:00",
      "updatedAt": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/swap_info/create

一次成交，对应合约事件 `SwapOnce`。`isBuy` 为 true 表示买入代币，`tokenAmount` 始终为正。`quoteAmount` 是扣费前的报价币，`fee` 从这笔报价币里扣出，单位与报价币相同。进出池子的报价币 = `quoteAmount - fee`。`price` 是这笔成交使用的曲线价格，不含本笔造成的变化。`chainId + txHash + logIndex` 唯一，重复写入返回 409。

请求：

```json
{
  "chainId": 1,
  "poolId": "pool-1",
  "txHash": "0xswap",
  "logIndex": 3,
  "trader": "0xtrader",
  "isBuy": true,
  "tokenAddr": "0xtoken",
  "tokenAmount": 1000,
  "tokenDecimal": 18,
  "quoteTokenAddr": "0xusdc",
  "quoteAmount": 200,
  "fee": 2,
  "quoteDecimal": 6,
  "price": 50
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "poolId": "pool-1",
    "txHash": "0xswap",
    "logIndex": 3,
    "trader": "0xtrader",
    "isBuy": true,
    "tokenAddr": "0xtoken",
    "tokenAmount": 1000,
    "tokenDecimal": 18,
    "quoteTokenAddr": "0xusdc",
    "quoteAmount": 200,
    "fee": 2,
    "quoteDecimal": 6,
    "price": 50,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/swap_info/update

请求：

```json
{
  "id": 1,
  "chainId": 1,
  "poolId": "pool-1",
  "txHash": "0xswap",
  "logIndex": 3,
  "trader": "0xtrader",
  "isBuy": false,
  "tokenAddr": "0xtoken",
  "tokenAmount": 1000,
  "tokenDecimal": 18,
  "quoteTokenAddr": "0xusdc",
  "quoteAmount": 180,
  "fee": 2,
  "quoteDecimal": 6,
  "price": 48
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "poolId": "pool-1",
    "txHash": "0xswap",
    "logIndex": 3,
    "trader": "0xtrader",
    "isBuy": false,
    "tokenAddr": "0xtoken",
    "tokenAmount": 1000,
    "tokenDecimal": 18,
    "quoteTokenAddr": "0xusdc",
    "quoteAmount": 180,
    "fee": 2,
    "quoteDecimal": 6,
    "price": 48,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T16:00:00+08:00"
  }
}
```

## POST /api/swap_info/delete

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

## GET /api/swap_info/get

请求：查询参数 `id=1`。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "poolId": "pool-1",
    "txHash": "0xswap",
    "logIndex": 3,
    "trader": "0xtrader",
    "isBuy": true,
    "tokenAddr": "0xtoken",
    "tokenAmount": 1000,
    "tokenDecimal": 18,
    "quoteTokenAddr": "0xusdc",
    "quoteAmount": 200,
    "fee": 2,
    "quoteDecimal": 6,
    "price": 50,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/swap_info/list

请求：查询参数 `chainId=1&poolId=pool-1&txHash=0xswap&trader=0xtrader&isBuy=true&tokenAddr=0xtoken&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainId": 1,
      "poolId": "pool-1",
      "txHash": "0xswap",
      "logIndex": 3,
      "trader": "0xtrader",
      "isBuy": true,
      "tokenAddr": "0xtoken",
      "tokenAmount": 1000,
      "tokenDecimal": 18,
      "quoteTokenAddr": "0xusdc",
      "quoteAmount": 200,
      "fee": 2,
      "quoteDecimal": 6,
      "price": 50,
      "createdAt": "2026-10-07T15:09:00+08:00",
      "updatedAt": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `id` 倒序。

## POST /api/sync_event/create

请求：

```json
{
  "chainId": 1,
  "blockNumber": 10,
  "blockHash": "0xblock",
  "txHash": "0xhash",
  "txIndex": 0,
  "logIndex": 2,
  "contractAddr": "0xcontract",
  "eventName": "TokenCreated",
  "topics": "0xtopic",
  "data": "0xdata",
  "removed": 0
}
```

`eventName` 是事件名称，例如 `TokenCreated`、`FeeCharged`、`SwapOnce`。扫块认不出的日志留空。`removed`：0 未移除，1 因链重组已移除。`chainId + txHash + logIndex` 唯一，重复写入返回 409。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "blockNumber": 10,
    "blockHash": "0xblock",
    "txHash": "0xhash",
    "txIndex": 0,
    "logIndex": 2,
    "contractAddr": "0xcontract",
    "eventName": "TokenCreated",
    "topics": "0xtopic",
    "data": "0xdata",
    "removed": 0,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/sync_event/update

请求：

```json
{
  "id": 1,
  "chainId": 1,
  "blockNumber": 10,
  "blockHash": "0xblock",
  "txHash": "0xhash",
  "txIndex": 0,
  "logIndex": 2,
  "contractAddr": "0xcontract",
  "eventName": "TokenCreated",
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
    "chainId": 1,
    "blockNumber": 10,
    "blockHash": "0xblock",
    "txHash": "0xhash",
    "txIndex": 0,
    "logIndex": 2,
    "contractAddr": "0xcontract",
    "eventName": "TokenCreated",
    "topics": "0xtopic",
    "data": "0xdata",
    "removed": 1,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T16:00:00+08:00"
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
    "chainId": 1,
    "blockNumber": 10,
    "blockHash": "0xblock",
    "txHash": "0xhash",
    "txIndex": 0,
    "logIndex": 2,
    "contractAddr": "0xcontract",
    "eventName": "TokenCreated",
    "topics": "0xtopic",
    "data": "0xdata",
    "removed": 0,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/sync_event/list

请求：查询参数 `chainId=1&txHash=0xhash&contractAddr=0xcontract&eventName=TokenCreated&removed=0&offset=0&limit=20`。筛选参数可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainId": 1,
      "blockNumber": 10,
      "blockHash": "0xblock",
      "txHash": "0xhash",
      "txIndex": 0,
      "logIndex": 2,
      "contractAddr": "0xcontract",
      "eventName": "TokenCreated",
      "topics": "0xtopic",
      "data": "0xdata",
      "removed": 0,
      "createdAt": "2026-10-07T15:09:00+08:00",
      "updatedAt": "2026-10-07T15:09:00+08:00"
    }
  ]
}
```

按 `blockNumber`、`logIndex` 倒序。

## POST /api/sync_cursor/create

请求：

```json
{
  "chainId": 1,
  "blockNumber": 100
}
```

`chainId` 唯一，同一条链重复创建返回 409。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "blockNumber": 100,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## POST /api/sync_cursor/update

请求：

```json
{
  "id": 1,
  "chainId": 1,
  "blockNumber": 250
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "chainId": 1,
    "blockNumber": 250,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T16:00:00+08:00"
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
    "chainId": 1,
    "blockNumber": 100,
    "createdAt": "2026-10-07T15:09:00+08:00",
    "updatedAt": "2026-10-07T15:09:00+08:00"
  }
}
```

## GET /api/sync_cursor/list

请求：查询参数 `chainId=1&offset=0&limit=20`。`chainId` 可不传。无 JSON 请求体。

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1,
      "chainId": 1,
      "blockNumber": 250,
      "createdAt": "2026-10-07T15:09:00+08:00",
      "updatedAt": "2026-10-07T16:00:00+08:00"
    }
  ]
}
```

按 `chainId` 升序。
