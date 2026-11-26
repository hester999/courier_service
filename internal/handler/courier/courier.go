package courier

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"avito/internal/apperr"
	"avito/internal/handler/courier/dto"
	"avito/internal/utils"
	"avito/pkg"

	"github.com/gorilla/mux"
)

type HandlerCourier struct {
	uc     usecasesCourier
	logger pkg.Logger
}

func NewHandlerCourier(uc usecasesCourier, logger pkg.Logger) *HandlerCourier {
	return &HandlerCourier{
		uc:     uc,
		logger: logger,
	}
}

const (
	lim = 10
	off = 0
)

func (h *HandlerCourier) GetCouriers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	ctx := r.Context()

	limit := int64(lim)
	offset := int64(off)

	if str := r.URL.Query().Get("limit"); str != "" {
		if l, err := strconv.ParseInt(str, 10, 64); err == nil {
			limit = l
		} else {
			utils.WriteError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
	}

	if str := r.URL.Query().Get("offset"); str != "" {
		if o, err := strconv.ParseInt(str, 10, 64); err == nil {
			offset = o
		} else {
			utils.WriteError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
	}

	if err := utils.ValidatePagination(&limit, &offset); err != nil {
		var appErr *apperr.AppErr
		if errors.As(err, &appErr) {
			utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	couriers, err := h.uc.GetCouriers(ctx, &limit, &offset)
	if err != nil {
		var appErr *apperr.AppErr
		if errors.As(err, &appErr) {
			switch {
			case appErr.Is(apperr.ErrNotFound):
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				utils.WriteError(w, http.StatusInternalServerError, appErr.Error())
				return
			}
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, dto.CouriersResponse(couriers))
}

func (h *HandlerCourier) GetCourier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	ctx := r.Context()

	var appErr *apperr.AppErr

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "id is must be integer")
		return
	}

	courier, err := h.uc.GetCourierById(ctx, id)
	if err != nil {
		if errors.As(err, &appErr) {
			switch {
			case appErr.Is(apperr.ErrNotFound):
				utils.WriteError(w, http.StatusNotFound, appErr.Error())
			default:
				utils.WriteError(w, http.StatusInternalServerError, "internal server error")
			}

			return
		}
	}

	response := dto.CourierResponse(courier)
	utils.WriteJSON(w, http.StatusOK, response)

}

func (h *HandlerCourier) CreateCourier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	ctx := r.Context()
	var appErr *apperr.AppErr

	var courier *dto.CreateCourierRequestDTO

	err := json.NewDecoder(r.Body).Decode(&courier)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.validateCreateParams(courier.Name, courier.Phone, courier.Status)
	if err != nil {
		if errors.As(err, &appErr) {
			utils.WriteError(w, http.StatusBadRequest, appErr.Error())
		}
		return
	}

	modelCourier := dto.CourierCreateRequestToModelCreateCourier(courier)
	res, err := h.uc.CreateCourier(ctx, modelCourier)
	if err != nil {
		if errors.As(err, &appErr) {
			switch {
			case appErr.Is(apperr.ErrAlreadyExists):
				utils.WriteError(w, http.StatusConflict, appErr.Error())
			case appErr.Is(apperr.ErrInvalidPhone):
				utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			case appErr.Is(apperr.ErrInvalidStatus):
				utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			case appErr.Is(apperr.ErrInvalidTransportType):
				utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			default:
				utils.WriteError(w, http.StatusInternalServerError, "internal error")
			}
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	response := dto.CourierResponse(res)
	utils.WriteJSON(w, http.StatusCreated, response)

}

func (h *HandlerCourier) UpdateCourier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	ctx := r.Context()

	var courier dto.CourierFullUpdateDTO

	err := json.NewDecoder(r.Body).Decode(&courier)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var appErr *apperr.AppErr

	model := dto.CourierFullUpdateRequestToModelFullUpdate(&courier)

	res, err := h.uc.UpdateCourier(ctx, model)
	if err != nil {
		if errors.As(err, &appErr) {
			switch {
			/*
				изначально проверял ошибку от бд о том что курьер с таким номером уже существует, так как в задании это PUT, и если id не передан
				или не существует, то должен запрос должен создавать запись, но так как в задании есть ответ 404,то пока остается как заглушка
			*/
			// case appErr.Is(apperr.ErrAlreadyExists):
			//	utils.WriteError(w, http.StatusConflict, appErr.Error())
			case appErr.Is(apperr.ErrInvalidStatus):
				utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			case appErr.Is(apperr.ErrInvalidPhone):
				utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			case appErr.Is(apperr.ErrNotFound):
				utils.WriteError(w, http.StatusNotFound, appErr.Error())
			case appErr.Is(apperr.ErrInvalidTransportType):
				utils.WriteError(w, http.StatusBadRequest, appErr.Error())
			default:
				utils.WriteError(w, http.StatusInternalServerError, "internal error")

			}
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	response := dto.CourierResponse(res)
	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *HandlerCourier) validateCreateParams(name, phone, status *string) error {

	if name == nil || *name == "" {
		err := &apperr.AppErr{
			Message:      "name is required",
			ServiceError: nil,
		}
		return err

	}
	if phone == nil || *phone == "" {
		err := &apperr.AppErr{
			Message:      "phone is required",
			ServiceError: nil,
		}
		return err
	}
	if status == nil || *status == "" {
		err := &apperr.AppErr{
			Message:      "status is required",
			ServiceError: nil,
		}
		return err
	}
	return nil
}
