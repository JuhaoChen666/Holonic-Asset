# Asset API Response Contract

本文档定义 Asset 详情接口返回给前端的数据结构。

推荐前端统一调用：

```http
GET /api/v1/asset/:asset_id
```

Asset 列表支持按名称/描述模糊搜索、tag 和 Asset type 筛选：

```http
GET /api/v1/projects/:project_id/assets?query=hero&tags=hero&tags=player&types=character
```

`query` 会不区分大小写地匹配 Asset 的 `name` 或 `description`。多个 `tags` 表示 Asset 必须同时包含这些 tag；多个 `types` 表示匹配其中任意一种类型。`query`、`tags`、`types` 都未提供时返回项目下全部 Asset。

更新 Asset 基本参数仍使用现有路由，但请求体不再局限于 tags：

```http
POST /api/v1/asset/tags
Content-Type: application/json
```

```json
{
  "assetId": 1001,
  "name": "Hero",
  "projectId": 101,
  "type": "character",
  "description": "Main character",
  "tags": ["player", "hero"],
  "attributes": {
    "rarity": "legendary"
  },
  "version": 2
}
```

更新接口支持部分字段更新；未传字段保持原值。`content` 不接受更新，避免修改 Asset 基本信息时覆盖原型、动画或 tileset 内容。返回值包含更新后的全部基本参数，不包含 `content`。

Asset 的完整业务内容放在 `data.content` 中，`data.type` 决定 `content` 的具体结构。

## Common Response

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "assetId": 1001,
    "name": "Hero",
    "projectId": 101,
    "type": "character",
    "description": "Main character",
    "tags": ["player", "hero"],
    "attributes": {},
    "content": {},
    "version": 1
  }
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `assetId` | `number` | Asset ID |
| `name` | `string` | Asset 名称 |
| `projectId` | `number` | 所属项目 ID |
| `type` | `string` | Asset 类型 |
| `attributes` | `object` | Asset 元数据 |
| `content` | `object` | Asset 完整业务内容 |
| `version` | `number` | Asset 业务版本 |

当前 Asset 类型：`character`、`object`、`tileSet`、`audio`、`ui`、`scenery`。

Asset 自身的元数据放在 `data.attributes`，Asset 业务内容中的扩展元数据放在 `data.content.metadata`。Task 的 ID、执行状态和错误信息属于 Task，不嵌套在 Asset Content 中；前端以 Asset 内容节点的 `status` 判断生成内容是否可用。

## Status

prototype、animation、direction、item 和 tile 都可以拥有以下状态：

```text
pending
processing
partial
completed
failed
cancelled
```

| 状态 | 前端处理方式 |
| --- | --- |
| `pending` | 任务尚未开始，继续轮询 |
| `processing` | 正在生成，继续轮询 |
| `partial` | 展示已完成内容并继续轮询 |
| `completed` | 内容完整，停止轮询 |
| `failed` | 停止轮询并展示错误 |
| `cancelled` | 停止轮询并提供重新生成入口 |

## Character

Character 包含 prototype 和 animations。animation 下的方向使用动态对象，方向数量不固定。

```json
{
  "viewMode": "top_down",
  "viewElements": ["up", "down", "left", "right"],
  "prototype": {
    "status": "completed",
    "directions": {
      "up": {
        "status": "completed",
        "image": {
          "url": "https://cdn.example.com/hero/prototype-up.png",
          "status": "completed"
        }
      },
      "down": {
        "status": "completed",
        "image": {
          "url": "https://cdn.example.com/hero/prototype-down.png",
          "status": "completed"
        }
      },
      "left": {
        "status": "completed",
        "image": {
          "url": "https://cdn.example.com/hero/prototype-left.png",
          "status": "completed"
        }
      },
      "right": {
        "status": "completed",
        "image": {
          "url": "https://cdn.example.com/hero/prototype-right.png",
          "status": "completed"
        }
      }
    }
  },
  "animations": [
    {
      "id": 3001,
      "name": "walk",
      "status": "partial",
      "directions": {
        "left": {
          "status": "completed",
          "frames": [
            {
              "url": "https://cdn.example.com/hero/walk/left/001.png",
              "status": "completed",
              "duration": 100
            }
          ]
        },
        "right": {
          "status": "processing"
        }
      }
    }
  ]
}
```

上面的四方向只是 `top_down` 的示例。方向集合由 `viewMode` 和 `viewElements` 决定，不固定为四个。

例如 `side_on` 可以只包含左右两个方向：

```json
{
  "viewMode": "side_on",
  "viewElements": ["left", "right"],
  "prototype": {
    "status": "completed",
    "directions": {
      "left": {
        "status": "completed",
        "image": {
          "id": 2101,
          "url": "https://cdn.example.com/hero/prototype-left.png",
          "status": "completed"
        }
      },
      "right": {
        "status": "completed",
        "image": {
          "id": 2102,
          "url": "https://cdn.example.com/hero/prototype-right.png",
          "status": "completed"
        }
      }
    }
  }
}
```

Character 约定：

- `viewMode` 表示 Asset 的视角模式，目前包括 `side_on` 和 `top_down`。
- `viewElements` 是当前 Asset 实际需要生成的方向集合。
- `prototype.directions` 的 key 必须与 `viewElements` 一一对应。
- 每个 prototype direction 在完成时必须且只能包含一张 `image`。
- `side_on` 可以是 `left`、`right` 两个方向，对应两张图片。
- `top_down` 可以是 `up`、`down`、`left`、`right` 四个方向，对应四张图片。
- `directions` 中的 key 表示本次实际选择生成的方向。
- 方向未完成时，可以不返回 `frames`。
- `partial` 表示部分方向已完成，前端可以先展示已完成方向。
- Character 创建后，prototype 初始状态为 `pending`。

## Object

Object 与 Character 使用相同的 content 结构，可以只有 prototype，也可以拥有动画。

```json
{
  "viewMode": "side_on",
  "viewElements": ["left", "right"],
  "prototype": {
    "status": "processing"
  },
  "animations": [
    {
      "id": 3101,
      "name": "open",
      "status": "pending",
      "directions": {
        "front": {
          "status": "pending"
        }
      }
    }
  ]
}
```

Object 的动画名称可以是 `open`、`close`、`destroyed`、`activated` 等业务动作。

## TileSet

TileSet 不包含 prototype 和 animations，而是由 item 列表组成，每个 item 下面包含 tile 列表。

```json
{
  "tileSize": {
    "width": 32,
    "height": 32
  },
  "items": [
    {
      "name": "grass",
      "status": "completed",
      "tiles": [
        {
          "name": "grass-center",
          "url": "https://cdn.example.com/tileset/grass/center.png",
          "status": "completed",
          "position": {
            "x": 0,
            "y": 1
          },
          "metadata": {
            "tileType": "center"
          }
        },
        {
          "name": "grass-top-left",
          "url": "https://cdn.example.com/tileset/grass/top-left.png",
          "status": "completed",
          "position": {
            "x": 1,
            "y": 1
          },
          "metadata": {
            "tileType": "top-left"
          }
        }
      ]
    },
    {
      "name": "water",
      "status": "processing"
    }
  ]
}
```

TileSet 约定：

- `tileSize` 定义整个 TileSet 的固定 tile 尺寸，所有 tile 使用相同的宽高。
- `items` 是 TileSet 的一级业务内容。
- `items[].tiles` 是该 item 生成出的 tile 列表。
- `tiles[].position` 是 tile 在网格中的位置，`x` 表示列、`y` 表示行，坐标从 `0` 开始；例如 `{ "x": 0, "y": 1 }` 表示第 0 列、第 1 行。
- item 可以独立处于 `pending`、`processing`、`partial` 或 `completed` 状态。
- 前端可以先展示已完成的 item，同时继续轮询其他 item。
- tile 的扩展属性放在 `metadata` 中，例如 tile 类型；固定尺寸和网格位置分别使用 `tileSize` 与 `position`。

## Polling

前端可以按照以下规则决定是否继续轮询：

```text
prototype.status in [pending, processing]
animation.status in [pending, processing, partial]
direction.status in [pending, processing]
item.status in [pending, processing, partial]
```

当目标节点为 `completed` 时，prototype direction 读取对应的单张 `image`，animation direction 读取 `frames`，item 读取 `tiles`。

当目标节点为 `failed` 时停止轮询；如果需要查看失败原因，查询对应的 Task 记录。
