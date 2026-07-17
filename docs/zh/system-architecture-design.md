# 系统架构设计

## 1.整体架构

### 1.1 架构风格
本项目采用 Service-Based Architecture（SBA）,根据业务领域与功能划分为几个独立的服务，每个服务都有明确的职责边界，可以根据业务需求选择相应的技术栈，并通过http/grpc等协议进行通信。

原因：
- 相较于单体架构，本项目业务量较大，不同服务领域区分清晰且有多种技术栈的需求，SBA提供了更高的技术栈灵活性与更好的模块解耦。
- 相较于微服务架构，本项目不需要引入复杂的服务治理基础设施，服务拆分粒度也不需要过细，架构简单，更适合当前项目规模。

![alt text](/docs/image/system-architecture.png)

### 1.2 拓扑关系

## 2. 服务内部层次

## 3. 业务服务设计

本节描述 Project、Asset 和 AI 三个业务服务的职责边界、内部模型及接口。

### 3.1 Project Service

#### 服务职责

Project Service 负责项目的生命周期及项目级配置管理，主要包括：

- 创建项目
- 获取当前用户的项目列表
- 获取项目详情
- 更新项目配置
- 维护项目类型、视角、美术风格、描述和参考图等项目级信息

Project Service 拥有 Project 相关数据。需要生成项目参考图时，由应用层调用 AI Service，Project 领域模型不直接依赖具体的模型供应商。

#### 领域模型及代码表示

`Project` 是项目领域的核心实体；`GameType` 和 `ViewType` 是描述项目属性的值类型。下面的 Go 定义是这些领域概念在代码中的表示，而不是独立的系统模块。

```go

type GameType string

type ViewType string



const (

GameTypeRPG GameType = "RPG"

GameTypeACT GameType = "ACT"

GameTypeSLG GameType = "SLG"

GameTypeOther GameType = "Other"

ViewTypeTopDown ViewType = "TopDown"

ViewTypeSideView ViewType = "SideView"

ViewTypeIsometric ViewType = "Isometric"

)



type Project struct {

ID uint

Name string

GameType GameType `json:"gameType"` // RPG、ACT、SLG 等

ViewType ViewType `json:"viewType"` // TopDown、SideView、Isometric 等

Description string // 项目描述

Reference string // 基于项目描述由 AI 生成的参考图

Style string // 项目的美术风格

}

```


#### 应用服务接口

`ProjectService` 定义项目相关用例，是接口层调用 Project 业务能力的应用层入口。

```go

type ProjectService interface {

Create(ctx context.Context, project *Project) error

// 根据用户 ID 获取项目列表

ListByUid(ctx context.Context, uid uint) ([]*Project, error)

// GetDetail 返回项目详情。

GetDetail(ctx context.Context, id uint) (*Project, error)

Update(ctx context.Context, project *Project) error

}
```

#### 对外接口

Project Service 对外提供以下业务能力：

- 创建项目
- 按用户查询项目列表
- 查询项目详情
- 更新项目配置

HTTP 路径、gRPC 方法、请求与响应 DTO、错误码及接口版本策略将在对外 API 设计中单独定义。接口层负责将外部 DTO 转换为应用服务所需的参数，不直接暴露领域对象或数据库模型。

### 3.2 Asset Service

#### 服务职责

Asset Service 负责项目内资产的生命周期、组织关系和版本管理，主要包括：

- 创建或复制一个或多个资产
- 按项目、类型、标签或名称查询资产
- 获取和更新资产详情
- 删除资产
- 建立资产之间的关联
- 管理资产标签
- 创建、查询和恢复资产历史版本

Asset Service 拥有 Asset、AssetResource、AssetSnapshot 和 AssetRecord 相关数据。其他服务不得绕过 Asset Service 直接修改这些数据。

#### 领域模型及代码表示

Asset 领域包含以下核心模型：

- `Asset`：资产当前可编辑状态
- `AssetResource`：资产依赖或引用的其他资源
- `AssetSnapshot`：资产在特定时间点的完整状态
- `AssetRecord`：不可变的资产历史版本
- `AssetType`：资产类型值对象

这些 Go 类型是 Asset 领域模型的代码表示。`AssetRecord` 与 `AssetSnapshot` 属于 Asset 领域内部的版本管理模型，不作为独立系统服务。
```go
type AssetType string

const (

AssetTypeCharacter AssetType = "character"

AssetTypeBackground AssetType = "background"

AssetTypeAudio AssetType = "audio"

AssetTypeUI AssetType = "UI"

AssetTypeObject AssetType = "object"

AssetTypeScenery AssetType = "scenery"

AssetTypeLayer AssetType = "layer"

)

// Asset 存储所有资产类型共有的字段。
// Attributes 使用 JSON 存储资产类型特有的信息，例如：
// - 画布信息
// - 动画信息
// - 音频元数据
// - 原型信息
// 服务层需要校验 Attributes 是否为合法的 JSON 对象。

type Asset struct {

ParentID unit `json:parentId`

ID uint `json:"id"`

ProjectID uint `json:"projectId"`

Name string `json:"name"`

Type AssetType `json:"type"`

Description string `json:"description"`

ResultURL string `json:"resultUrl"`

Tags []string `json:"tags"`

Attributes json.RawMessage `json:"attributes"`

}

// AssetResource 表示当前资产所依赖或引用的其他资产。
// 资源信息会保存在快照中，以确保历史版本能够保留当时的依赖关系。

type AssetResource struct {

AssetID uint `json:"assetId"`

Name string `json:"name"`

URL string `json:"url"`

}

// AssetSnapshot 表示资产在某个时间点的完整可编辑状态。

// ID 和 ProjectID 会被保留用于审计，但恢复快照时不得修改

// 当前资产的身份标识或所属项目。

type AssetSnapshot struct {

Asset Asset `json:"asset"`

Resources []AssetResource `json:"resources,omitempty"`

Attributes json.RawMessage `json:"attributes"`

}

// AssetRecord 表示一个不可变的资产历史版本。
// Snapshot 在数据库中以 JSON 格式存储。
// AssetSnapshot 定义了序列化和读取快照时所使用的文档结构。

type AssetRecord struct {
ID uint `json:"id"`
AssetVersion uint `json:"assetVersion"`
AssetID uint `json:"assetId"`
Snapshot json.RawMessage `json:"snapshot"`

}

```

#### 应用服务接口

`AssetService` 提供资产当前状态的管理用例，`AssetRecordService` 提供资产版本管理用例。

```go

type AssetService interface {

// Create 创建资产，并同时生成该资产的初始版本快照。

Create(ctx context.Context, asset *Asset) error

// ListByProjectID 返回指定项目下的所有资产。

ListByProjectID(ctx context.Context, projectID uint) ([]*Asset, error)

// GetDetail 返回指定资产的当前详细信息。

GetDetail(ctx context.Context, id uint) (*Asset, error)
// Update 更新资产，并在同一个事务中创建新的版本快照。

Update(ctx context.Context, asset *Asset) error
}

type AssetRecordService interface {

// CreateSnapshot 根据资产的当前状态创建快照。
// 具体的 AssetVersion 由服务层自动计算和分配。

CreateSnapshot(ctx context.Context, assetID uint) (*AssetRecord, error)

// ListByAssetID 返回指定资产的所有快照记录，
// 并按照 AssetVersion 从高到低排序。

ListByAssetID(
ctx context.Context,
assetID uint,
) ([]*AssetRecord, error)

// GetDetail 返回指定资产快照记录的详细信息。

GetDetail(ctx context.Context, recordID uint) (*AssetRecord, error)

// Restore 使用指定快照恢复资产的可编辑状态。
// 恢复操作会创建一个新的资产版本，不会覆盖或删除已有历史记录。

Restore(ctx context.Context, assetID uint, recordID uint,
) (*AssetRecord, error)
}

```

当前接口已经覆盖创建、列表查询、详情查询、更新、快照创建、快照查询和版本恢复。删除、搜索、资产关联和标签管理等能力仍需在应用服务接口中补充。

#### 对外接口

Asset Service 对外提供资产管理和Record两类接口。

资产管理接口包括创建、复制、查询、更新、删除、搜索、关联和标签管理；版本管理接口包括查询版本列表、获取版本详情和恢复指定版本。

具体协议、API 路径、分页规则、筛选参数及错误码将在对外 API 设计中定义。接口层不得允许调用方通过快照恢复操作修改资产 ID 或所属项目。

### 3.3 AI Service

#### 服务职责

AI Service 为其他业务服务和外部调用方提供内容生成能力，主要包括：

- 生成人物资产
- 生成 UI 元素
- 生成场景和图层
- 生成图块集
- 生成对象
- 生成动画
- 生成项目参考图

AI Service 负责生成任务的组织和模型调用，不拥有 Project 或 Asset 的业务数据。生成结果需要保存为资产时，应由相应的应用服务协调 Asset Service 完成。

#### 应用接口 DTO

`CreateCharacterRequest`、`CreateSceneRequest` 等类型描述 AI 生成用例的输入与输出。它们是应用接口 DTO，不是具有业务身份和生命周期的领域实体。

```go

type Size struct {
Width int `json:"width"`
Height int `json:"height"`
}

type CreateCharacterRequest struct {
ProjectPrompt string `json:"projectPrompt"` // 项目提示词
UserPrompt string `json:"userPrompt"`
Name string `json:"name"`
Facing string `json:"facing"`
Size Size `json:"size"`
Reference []string `json:"reference"`
Physics PhysicsConfig `json:"physics"`
}

type CreateCharacterResponse struct {
URL string `json:"url"`
}

type PhysicsConfig struct {
Collision CollisionConfig `json:"collision"`
Movement MovementConfig `json:"movement"`
Gravity GravityConfig `json:"gravity"`
}

type CreateUIRequest struct {
ProjectPrompt string `json:"projectPrompt"` // 项目提示词
UserPrompt string `json:"user_prompt"`
Type string `json:"type"` // button、panel、hp_bar
Size Size `json:"size"`
Reference []string `json:"reference"`
}

type CreateUIResponse struct {
URL string `json:"url"`
}

type LayerResult struct {
ID uint `json:"id"` // 图层 ID
Url string `json:"url"` // 生成图片的 URL
}

type CreateSceneRequest struct {
ProjectPrompt string `json:"projectPrompt"` // 项目提示词
Style string `json:"style"` // 场景风格
Layers []Layer `json:"layers"` // 场景图层
}

type CreateSceneResponse struct {
Layers []LayerResult `json:"layers"` // 每个图层的生成结果
}

type CreateTileSetRequest struct {
ProjectPrompt string `json:"projectPrompt"` // 项目提示词
Prompt string `json:"prompt"` // 图块集提示词
Reference []string `json:"reference"` // 创建图块集所用的参考图
}

type CreateTileSetResponse struct {
Url string `json:"url"` // 生成图块集图片的 URL
}

type CreateObjectRequest struct {
UserPrompt string `json:"prompt"` // 对象提示词
ProjectPrompt string `json:"projectPrompt"` // 项目提示词
Derictions int `json:"derictions"` // 对象方向数量，例如 1、4、8
Reference string `json:"reference"` // 创建对象所用的参考图
Size Size `json:"size"` // 对象尺寸，例如 "32X32"、"64X64"
View ViewType `json:"view"` // 对象视角，例如 "TopDown"、"SideView"、"Isometric"
}

type CreateObjectResponse struct {
Url string `json:"url"` // 生成对象图片的 URL
}

type CreateAnimationRequest struct {
ProjectPrompt string `json:"projectPrompt"`
UserPrompt string `json:"userPrompt"`
Name string `json:"name"`
FirstFrameURL string `json:"firstFrameUrl"`
Description string `json:"description"`
FrameCount int `json:"frameCount"`
KeepFirstFrame bool `json:"keepFirstFrame"`
}

type CreateAnimationResponse struct {
URL string `json:"urls"`
}

```

DTO 负责表达生成参数和结果，不应包含具体模型供应商的私有协议。供应商参数应在基础设施适配层完成转换。

#### 应用服务接口

AI 应用服务接口用于承载人物、场景、图块集、对象、UI 和动画等生成用例。

```go

type CharacterService interface {
CrreateCharacter(request *CreateCharacterRequest)
}

type MapService interface {
CreateScene(request *CreateSceneRequest) (*CreateSceneResponse, error)
CreateTileSet(request *CreateTileSetRequest) (*CreateTileSetResponse, error)
}

type ObjectService interface {
CreateObject(request *CreateObjectRequest) (*CreateObjectResponse, error)
}
```

当前接口仅覆盖人物、场景、图块集和对象生成。UI、动画及参考图生成接口仍需补充。应用服务接口应统一接收 `context.Context`，以支持超时、取消和请求追踪。

#### 模型供应商适配接口

`LLMMessage`、`LLMRequest`、`ImageGenerationRequest` 等类型用于描述 AI Service 与模型供应商之间的交互协议。

`LLMClient` 是 AI Service 调用外部模型能力的端口，不属于 Project、Asset 等业务领域模型。具体供应商客户端在基础设施层实现该接口。

```go
type MessageRole string
type ContentPartType string

const (
MessageRoleSystem MessageRole = "system"
MessageRoleUser MessageRole = "user"
MessageRoleAssistant MessageRole = "assistant"
MessageRoleTool MessageRole = "tool"
ContentPartText ContentPartType = "text"
ContentPartImageURL ContentPartType = "image_url"
ContentPartAudioURL ContentPartType = "audio_url"
ContentPartMaskURL ContentPartType = "mask_url"
)



type ContentPart struct {
Type ContentPartType `json:"type"`
Text string `json:"text,omitempty"`
URL string `json:"url,omitempty"`
MediaType string `json:"mediaType,omitempty"`
}



type LLMMessage struct {
Role MessageRole `json:"role"`
Content []ContentPart `json:"content"`
}



type LLMUsage struct {
InputTokens int `json:"inputTokens"`
OutputTokens int `json:"outputTokens"`
TotalTokens int `json:"totalTokens"`
}



type LLMRequest struct {
RequestID string `json:"requestId"`
Model string `json:"model"`
Messages []LLMMessage `json:"messages"`
ResponseFormat json.RawMessage `json:"responseFormat,omitempty"`
}



type LLMResponse struct {
ID string `json:"id"`
Model string `json:"model"`
Message LLMMessage `json:"message"`
Usage LLMUsage `json:"usage"`
}

type ImageGenerationRequest struct {
RequestID string `json:"requestId"`
Model string `json:"model"`
Prompt string `json:"prompt"`
References []string `json:"references,omitempty"`
Size Size `json:"size"`
Count int `json:"count"`
}

type LLMClient interface {
Chat(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
GenerateImage(ctx context.Context, request *ImageGenerationRequest) (*GenerationResult, error)
GetGenerationResult(ctx context.Context, generationID string) (*GenerationResult, error)
CancelGeneration(ctx context.Context, generationID string) error
}
```

供应商适配层负责：

- 将应用接口 DTO 转换为供应商请求
- 发起文本或图像生成任务
- 查询和取消异步生成任务
- 将供应商响应转换为统一结果
- 隔离不同模型供应商的协议差异

## 4. 接入基础设施

### 4.1 Gateway 职责

Gateway 为前端和外部调用方提供统一的系统入口。系统初期使用 Nginx 作为 Gateway 的实现。
Gateway 负责：

- 将请求转发到对应的后端服务
- 终止 TLS 连接
- 转发用户认证信息
- 处理跨域请求
- 限制请求体大小
- 实施请求限流
- 控制请求超时时间
- 记录访问日志和请求追踪信息

### 4.2 边界约束

Gateway 属于接入基础设施，不是业务服务，也不包含 Project、Asset 或 AI 领域逻辑。
Gateway 不得：

- 直接访问任何业务服务的数据库
- 执行资产、项目或生成任务的业务规则
- 直接调用外部模型供应商
- 根据业务状态决定跨服务流程
- 修改业务服务返回的数据语义
Gateway 可以完成认证信息转发，但具体资源访问权限仍由对应业务服务校验。

### 4.3 请求路由与服务编排
