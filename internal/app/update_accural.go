package app

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"gophermart/internal/service"
	"gophermart/internal/storage"
	"log"
	"net/http"
	"time"
)

func (app *App) UpdateAccrual() error {
	ordersToUpdate, err := app.userStorage.GetOrdersToUpdate()
	if err != nil {
		return err
	}

	if len(ordersToUpdate) != 0 {
		req := resty.New().
			SetBaseURL(app.config.AccrualAddress).
			R().
			SetHeader("Content-Type", "application/json")
		for _, order := range ordersToUpdate {

			orderNum := order.Number
			resp, err := req.Get("/api/orders/" + orderNum)
			if err != nil {
				return err
			}

			status := resp.StatusCode()
			switch status {
			case http.StatusTooManyRequests:
				time.Sleep(storage.UPDATEACCURALTIMER)
				return nil

			case http.StatusOK:
				var updatedOrder service.AccrualResponse
				err = json.Unmarshal(resp.Body(), &updatedOrder)
				if err != nil {
					log.Printf("json decode order accrual: %s", err)
					return err
				}
				log.Printf("accrual for order %s updating to %s", updatedOrder.OrderID, updatedOrder.Status)

				orderToUpload := service.Order{
					Number:  updatedOrder.OrderID,
					Login:   order.Login,
					Status:  updatedOrder.Status,
					Accrual: updatedOrder.Accrual,
				}
				err = app.userStorage.UpdateOrderStatus(orderToUpload)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
