type CharacterService interface {
	CrreateCharacter(
		ctx context.Context,
		request *CreateCharacterRequest)(*CreateCharacterResponse, error)

	EditCharacter(
		ctx context.Context,
		request *CreateCharacterRequest)(*CreateCharacterResponse, error)
}

type SceneryService interface{
	CreateLayer(
		ctx context.Context,
		request *CreateLayerRequest) (*CreateLayerResponse, error)

	EditLayer(
		ctx context.Context,
		request *CreateLayerRequest) (*CreateLayerResponse, error)
}

type TileSetService interface{
	CreateItem(
		ctx context.Context,
		request *CreateItemRequest) (*CreateItemResponse, error)

	EditItem(
		ctx context.Context,
		request *CreateItemRequest) (*CreateItemResponse, error)
}

type ObjectService interface {
	CreateObject(
		ctx context.Context,
		request *CreateObjectRequest) (*CreateObjectResponse, error)

	EditObject(
		ctx context.Context,
		request *CreateObjectRequest) (*CreateObjectResponse, error)
}

type ProjectService interface{
	CreateProject(
		ctx context.Context,
		request *CreateProjectRequest) (*CreateProjectResponse, error)
}

type AnimationService interface{
	CreateAnimation(
		ctx context.Context,
		request *CreateAnimationRequest) (*CreateAnimationResponse, error)

	EditFrame(
		ctx context.Context,
		request *CreateFrameRequest) (*CreateFrameResponse, error)
}

type UIService interface{
	CreateUI(
		ctx context.Context,
		request *CreateUIRequest) (*CreateUIResponse, error)

	EditUIcompoent(
		ctx context.Context,
		request *CreateUIcompoentRequest) (*CreateUIcompoentResponse, error)
}
