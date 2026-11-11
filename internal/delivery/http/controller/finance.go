package controller

//go:generate go tool mockgen -source=finance.go -destination=../../../../mock/controller_finance.go -package=mock

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	ListTransactions(ctx context.Context, req domain.ListTransactionsRequest) ([]domain.Transaction, error)
	CreateTransaction(ctx context.Context, req domain.SaveTransactionRequest) ([]domain.Transaction, error)
	UpdateTransaction(ctx context.Context, req domain.SaveTransactionRequest) ([]domain.Transaction, error)
	DeleteTransaction(ctx context.Context, req domain.DeleteTransactionRequest) error
	ListRevisions(ctx context.Context, id int64) ([]domain.TransactionRevision, error)
	SaveReceipt(ctx context.Context, id int64, data []byte) error
	GetReceipt(ctx context.Context, id int64) (domain.TransactionPhoto, error)
	MoveFunds(ctx context.Context, req domain.MoveFundsRequest) ([]domain.Transaction, error)

	Summary(ctx context.Context, month string) (domain.FinanceSummary, error)
	FinishPocket(ctx context.Context, req domain.FinishPocketRequest) (domain.Pocket, error)
	ListClosings(ctx context.Context) ([]domain.BookClosing, error)
	CloseMonth(ctx context.Context, req domain.CloseMonthRequest) (domain.BookClosing, error)
	GetTVSetting(ctx context.Context) (domain.FinanceTVSetting, error)
	UpdateTVSetting(ctx context.Context, req domain.UpdateFinanceTVSettingRequest) (domain.FinanceTVSetting, error)
	GetDisplay(ctx context.Context) (domain.FinanceDisplay, error)
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

func (ctrl *Controller) ListTransactions(c echo.Context) error {
	var req domain.ListTransactionsRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	return respond(c, http.StatusOK)(ctrl.finance.ListTransactions(c.Request().Context(), req))
}

func (ctrl *Controller) CreateTransaction(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.SaveTransactionRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UserID = session.UserID
	return respond(c, http.StatusCreated)(ctrl.finance.CreateTransaction(c.Request().Context(), req))
}

func (ctrl *Controller) UpdateTransaction(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	id, err := paramID(c)
	if err != nil {
		return err
	}
	var req domain.SaveTransactionRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UserID, req.ID = session.UserID, id
	return respond(c, http.StatusOK)(ctrl.finance.UpdateTransaction(c.Request().Context(), req))
}

func (ctrl *Controller) DeleteTransaction(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	id, err := paramID(c)
	if err != nil {
		return err
	}
	var req domain.DeleteTransactionRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UserID, req.ID = session.UserID, id
	if err := ctrl.finance.DeleteTransaction(c.Request().Context(), req); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, messageResponse{Message: "transaksi dihapus"})
}

func (ctrl *Controller) ListRevisions(c echo.Context) error {
	return byID(c, ctrl.finance.ListRevisions)
}

func (ctrl *Controller) SaveReceipt(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	header, err := c.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
			return apperror.New(apperror.CodeInvalidInput, err, "foto bukti wajib diunggah")
		}
		return errBadRequest(err)
	}
	file, err := header.Open()
	if err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error opening receipt: %w", err), "gagal membaca foto")
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, domain.MaxReceiptPhotoSize+1))
	if err != nil {
		return apperror.New(apperror.CodeInternal, fmt.Errorf("error reading receipt: %w", err), "gagal membaca foto")
	}
	if err := ctrl.finance.SaveReceipt(c.Request().Context(), id, data); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, messageResponse{Message: "foto bukti disimpan"})
}

func (ctrl *Controller) GetReceipt(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}
	photo, err := ctrl.finance.GetReceipt(c.Request().Context(), id)
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderCacheControl, "private, max-age=300")
	return c.Blob(http.StatusOK, photo.ContentType, photo.Data)
}

func (ctrl *Controller) MoveFunds(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.MoveFundsRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UserID, req.GroupID = session.UserID, c.Param("group_id")
	status := http.StatusCreated
	if req.GroupID != "" {
		status = http.StatusOK
	}
	return respond(c, status)(ctrl.finance.MoveFunds(c.Request().Context(), req))
}

func (ctrl *Controller) GetFinanceSummary(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.Summary(c.Request().Context(), c.QueryParam("month")))
}

func (ctrl *Controller) FinishPocket(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	id, err := paramID(c)
	if err != nil {
		return err
	}
	var req domain.FinishPocketRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UserID, req.ID = session.UserID, id
	return respond(c, http.StatusOK)(ctrl.finance.FinishPocket(c.Request().Context(), req))
}

func (ctrl *Controller) ListClosings(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.ListClosings(c.Request().Context()))
}

func (ctrl *Controller) CloseMonth(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.CloseMonthRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UserID = session.UserID
	return respond(c, http.StatusCreated)(ctrl.finance.CloseMonth(c.Request().Context(), req))
}

func (ctrl *Controller) GetFinanceTVSetting(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.GetTVSetting(c.Request().Context()))
}

func (ctrl *Controller) UpdateFinanceTVSetting(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}
	var req domain.UpdateFinanceTVSettingRequest
	if err := c.Bind(&req); err != nil {
		return errBadRequest(err)
	}
	req.UpdatedBy = session.UserID
	return respond(c, http.StatusOK)(ctrl.finance.UpdateTVSetting(c.Request().Context(), req))
}

// GetFinanceDisplay is the TV's kas header
func (ctrl *Controller) GetFinanceDisplay(c echo.Context) error {
	return respond(c, http.StatusOK)(ctrl.finance.GetDisplay(c.Request().Context()))
}
