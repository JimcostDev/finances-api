package services

import (
	"context"

	"github.com/JimcostDev/finances-api/models"
	"github.com/JimcostDev/finances-api/repositories"
)

type CategoryService interface {
	GetCategories(ctx context.Context) ([]models.Category, error)
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) GetCategories(ctx context.Context) ([]models.Category, error) {
	return s.repo.FindAll(ctx)
}

