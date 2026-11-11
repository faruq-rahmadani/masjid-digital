package controller

//go:generate go tool mockgen -source=finance.go -destination=../../../../mock/controller_finance.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// FinanceUsecaseItf is Keuangan (E7): kas, akun, kategori and pembukuan
type FinanceUsecaseItf interface {
	ListAccounts(ctx context.Context) ([]domain.AccountWithBalance, error)
	CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (domain.Account, error)
	UpdateAccount(ctx context.Context, req domain.UpdateAccountRequest) (domain.Account, error)
	ArchiveAccount(ctx context.Context, id int64) (domain.Account, error)
	RestoreAccount(ctx context.Context, id int64) (domain.Account, error)

	ListPockets(ctx context.Context) ([]domain.PocketWithBalance, error)
	CreatePocket(ctx context.Context, req domain.CreatePocketRequest) (domain.Pocket, error)
	UpdatePocket(ctx context.Context, req domain.UpdatePocketRequest) (domain.Pocket, error)
	ArchivePocket(ctx context.Context, id int64) (domain.Pocket, error)
	RestorePocket(ctx context.Context, id int64) (domain.Pocket, error)

	ListCategories(ctx context.Context) ([]domain.CategoryWithCount, error)
	CreateCategory(ctx context.Context, req domain.CreateCategoryRequest) (domain.Category, error)
	UpdateCategory(ctx context.Context, req domain.UpdateCategoryRequest) (domain.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
	ArchiveCategory(ctx context.Context, id int64) (domain.Category, error)
	RestoreCategory(ctx context.Context, id int64) (domain.Category, error)

	GetBookkeeping(ctx context.Context) (domain.Bookkeeping, error)
	StartBookkeeping(ctx context.Context, req domain.StartBookkeepingRequest) (domain.Bookkeeping, error)
}

func (ctrl *Controller) ListAccounts(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.ListAccounts(c.Request().Context()))
}

func (ctrl *Controller) CreateAccount(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.CreatedBy = session.UserID
	return respond(c, http.StatusCreated)(ctrl.finance.CreateAccount(c.Request().Context(), req))
}

func (ctrl *Controller) UpdateAccount(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	var req domain.UpdateAccountRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.ID = id
	return respond(c, http.StatusOK)(ctrl.finance.UpdateAccount(c.Request().Context(), req))
}

func (ctrl *Controller) ArchiveAccount(c echo.Context) error {
	return byID(c, ctrl.finance.ArchiveAccount)
}

func (ctrl *Controller) RestoreAccount(c echo.Context) error {
	return byID(c, ctrl.finance.RestoreAccount)
}

func (ctrl *Controller) ListPockets(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.ListPockets(c.Request().Context()))
}

func (ctrl *Controller) CreatePocket(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.CreatePocketRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.CreatedBy = session.UserID
	return respond(c, http.StatusCreated)(ctrl.finance.CreatePocket(c.Request().Context(), req))
}

func (ctrl *Controller) UpdatePocket(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	var req domain.UpdatePocketRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.ID = id
	return respond(c, http.StatusOK)(ctrl.finance.UpdatePocket(c.Request().Context(), req))
}

func (ctrl *Controller) ArchivePocket(c echo.Context) error {
	return byID(c, ctrl.finance.ArchivePocket)
}

func (ctrl *Controller) RestorePocket(c echo.Context) error {
	return byID(c, ctrl.finance.RestorePocket)
}

func (ctrl *Controller) ListCategories(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.ListCategories(c.Request().Context()))
}

func (ctrl *Controller) CreateCategory(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.CreatedBy = session.UserID
	return respond(c, http.StatusCreated)(ctrl.finance.CreateCategory(c.Request().Context(), req))
}

func (ctrl *Controller) UpdateCategory(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	var req domain.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.ID = id
	return respond(c, http.StatusOK)(ctrl.finance.UpdateCategory(c.Request().Context(), req))
}

func (ctrl *Controller) DeleteCategory(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	if err := ctrl.finance.DeleteCategory(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, messageResponse{Message: "kategori dihapus"})
}

func (ctrl *Controller) ArchiveCategory(c echo.Context) error {
	return byID(c, ctrl.finance.ArchiveCategory)
}

func (ctrl *Controller) RestoreCategory(c echo.Context) error {
	return byID(c, ctrl.finance.RestoreCategory)
}

func (ctrl *Controller) GetBookkeeping(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.GetBookkeeping(c.Request().Context()))
}

func (ctrl *Controller) StartBookkeeping(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.StartBookkeepingRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.ClosedBy = session.UserID
	return respond(c, http.StatusCreated)(ctrl.finance.StartBookkeeping(c.Request().Context(), req))
}

// respond writes a usecase result as {"data": ...} with status, or returns its error
func respond(c echo.Context, status int) func(any, error) error {
	return func(data any, err error) error {
		if err != nil {
			return err
		}
		return c.JSON(status, dataResponse{Data: data})
	}
}

// byID runs a usecase action on the :id path parameter and writes its result
func byID[T any](c echo.Context, action func(context.Context, int64) (T, error)) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	return respond(c, http.StatusOK)(action(c.Request().Context(), id))
}

func errBadRequest(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
}
