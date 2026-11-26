package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"avito/internal/apperr"
	"avito/internal/handler/delivery/dto"
	"avito/internal/utils"
	"avito/pkg"
)

type DeliveryHandler struct {
	uc  deliveryUsecase
	log pkg.Logger
}

func NewDeliveryHandler(uc deliveryUsecase, log pkg.Logger) *DeliveryHandler {
	return &DeliveryHandler{uc: uc, log: log}
}

func (d *DeliveryHandler) AssignCourierOnOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req dto.AssignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OrderID == nil {
		utils.WriteError(w, http.StatusBadRequest, "empty orderID")
		return
	}

	res, err := d.uc.AssignCourier(r.Context(), dto.RequestAssignToModel(req))

	if err != nil {
		var appErr *apperr.AppErr
		d.log.Error("AssignCourierOnOrder: %v", err)
		if errors.As(err, &appErr) {
			switch {
			case appErr.Is(apperr.ErrAlreadyExists):
				utils.WriteError(w, http.StatusConflict, appErr.Message)
				return
			case appErr.Is(apperr.ErrAvailableCouriers):
				utils.WriteError(w, http.StatusConflict, appErr.Message)
				return
			case appErr.Is(apperr.ErrNotFound):
				utils.WriteError(w, http.StatusNotFound, "not found order")
				return
			default:
				utils.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response := dto.ModelToResponseAssign(res)
	utils.WriteJSON(w, http.StatusOK, response)
}

func (d *DeliveryHandler) UnassignCourierOnOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var req dto.AssignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OrderID == nil {
		utils.WriteError(w, http.StatusBadRequest, "empty orderID")
		return
	}
	res, err := d.uc.UnassignCourier(r.Context(), dto.RequestAssignToModel(req))
	if err != nil {
		var appErr *apperr.AppErr
		d.log.Error("UnassignCourierOnOrder: %v", err)
		if errors.As(err, &appErr) {
			switch {
			case appErr.Is(apperr.ErrNotFound):
				utils.WriteError(w, http.StatusNotFound, appErr.Message)
				return
			default:
				utils.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}
		}
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return

	}
	response := dto.ModelToResponseUnassign(res)
	utils.WriteJSON(w, http.StatusOK, response)
}
