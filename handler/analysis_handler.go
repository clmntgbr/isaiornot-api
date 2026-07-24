package handler

import (
	"errors"
	"go-api/handler/context"
	mediadto "go-api/infrastructure/media"
	"go-api/infrastructure/paginate"
	"go-api/presenter"
	"go-api/usecase/analysis"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AnalysisHandler struct {
	generatePresignedUploadUrlUseCase *analysis.GeneratePresignedUploadUrlUseCase
	getAnalysisUseCase                *analysis.GetAnalysisUseCase
	getAnalysesUseCase                *analysis.GetAnalysesUseCase
	getStatisticsUseCase              *analysis.GetStatisticsUseCase
}

func NewAnalysisHandler(
	generatePresignedUploadUrlUseCase *analysis.GeneratePresignedUploadUrlUseCase,
	getAnalysisUseCase *analysis.GetAnalysisUseCase,
	getAnalysesUseCase *analysis.GetAnalysesUseCase,
	getStatisticsUseCase *analysis.GetStatisticsUseCase,
) *AnalysisHandler {
	return &AnalysisHandler{
		generatePresignedUploadUrlUseCase: generatePresignedUploadUrlUseCase,
		getAnalysisUseCase:                getAnalysisUseCase,
		getAnalysesUseCase:                getAnalysesUseCase,
		getStatisticsUseCase:              getStatisticsUseCase,
	}
}

func (h *AnalysisHandler) GeneratePresignedUploadUrl(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var request mediadto.PresignUploadInput
	if err := c.Bind().JSON(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"errors":  err.Error(),
		})
	}

	result, err := h.generatePresignedUploadUrlUseCase.Execute(c.Context(), user.ID, request)
	if err != nil {
		if errors.Is(err, analysis.ErrUnsupportedMediaType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unsupported media type",
				"errors":  err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(presenter.NewGeneratePresignedUploadUrlDetailResponse(result.URL, result.AnalysisID.String()))
}

func (h *AnalysisHandler) GetAnalyses(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	query.Normalize()

	analyses, total, err := h.getAnalysesUseCase.Execute(c.Context(), user.ID, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(paginate.NewPaginateResponse(presenter.NewAnalysisListResponses(analyses), int(total), query))
}

func (h *AnalysisHandler) GetAnalysis(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	analysisID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	result, err := h.getAnalysisUseCase.Execute(c.Context(), user.ID, analysisID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Analysis not found",
		})
	}

	return c.JSON(presenter.NewAnalysisDetailResponse(result))
}

func (h *AnalysisHandler) GetStatistics(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	stats, err := h.getStatisticsUseCase.Execute(c.Context(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(presenter.NewMediaStatisticsResponse(stats))
}
