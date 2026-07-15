package services

type RepoService interface{}
type RepoServiceImpl struct{}

func NewRepoService() RepoService {
	return &RepoServiceImpl{}
}
