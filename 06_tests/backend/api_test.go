// api_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Helper: initialize Echo with in-memory SQLite and all routes
func setupAPIServer() (*echo.Echo, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&Product{}, &Cart{}, &Category{})

	e := echo.New()
	e.Use(DBMiddleware(db))

	// Raw echo endpoint
	e.POST("/test", func(c echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	})

	// Product endpoints
	e.POST("/products", createProduct)
	e.GET("/products/:id", getProduct)
	e.GET("/products", getAllProducts)
	e.PUT("/products/:id", updateProduct)
	e.DELETE("/products/:id", deleteProduct)

	// Cart endpoints
	e.POST("/carts", createCart)
	e.POST("/carts/:id/products", addProductToCart)
	e.GET("/carts/:id", getCart)

	// Category endpoints
	e.POST("/categories", createCategory)
	e.GET("/categories/:id", getCategory)

	return e, db
}

//--------------------------------------
// API tests
//--------------------------------------

// TestRawEndpoint covers POST /test
func TestRawEndpoint_API(t *testing.T) {
	e, _ := setupAPIServer()
	payload := `{"hello":"world"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, payload, rec.Body.String())

	// Negative: wrong content type
	reqBad := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(payload)))
	reqBad.Header.Set(echo.HeaderContentType, "text/plain")
	recBad := httptest.NewRecorder()
	e.ServeHTTP(recBad, reqBad)
	assert.Equal(t, http.StatusOK, recBad.Code) // still echoes body as string
}

//--------------------------------------
// Product API tests
//--------------------------------------

func TestProductEndpoints_API(t *testing.T) {
	e, _ := setupAPIServer()

	// Create a category for product
	catPayload := map[string]interface{}{"name": "Gadgets"}
	catBody, _ := json.Marshal(catPayload)
	reqCat := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(catBody))
	reqCat.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recCat := httptest.NewRecorder()
	e.ServeHTTP(recCat, reqCat)
	assert.Equal(t, http.StatusCreated, recCat.Code)

	var catResp CategoryResponse
	json.Unmarshal(recCat.Body.Bytes(), &catResp)

	// 1) POST /products (valid)
	prodPayload := map[string]interface{}{
		"name":         "Smartphone",
		"description":  "Android phone",
		"price":        699.99,
		"category_id":  catResp.ID,
	}
	prodBody, _ := json.Marshal(prodPayload)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(prodBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var createdProd ProductResponse
	json.Unmarshal(rec.Body.Bytes(), &createdProd)
	assert.Equal(t, "Smartphone", createdProd.Name)

	// Negative: POST /products with invalid JSON
	reqInv := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte(`{"name":123}`)))
	reqInv.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recInv := httptest.NewRecorder()
	e.ServeHTTP(recInv, reqInv)
	assert.Equal(t, http.StatusBadRequest, recInv.Code)

	// 2) GET /products/:id (valid)
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%d", createdProd.ID), nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	assert.Equal(t, http.StatusOK, recGet.Code)

	var fetchedProd ProductResponse
	json.Unmarshal(recGet.Body.Bytes(), &fetchedProd)
	assert.Equal(t, createdProd.ID, fetchedProd.ID)

	// Negative: GET /products/:id with non-numeric ID
	reqGetBad := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
	recGetBad := httptest.NewRecorder()
	e.ServeHTTP(recGetBad, reqGetBad)
	assert.Equal(t, http.StatusOK, recGetBad.Code)
	var emptyProd ProductResponse
	json.Unmarshal(recGetBad.Body.Bytes(), &emptyProd)
	assert.Equal(t, uint(0), emptyProd.ID)

	// 3) GET /products (valid)
	reqAll := httptest.NewRequest(http.MethodGet, "/products", nil)
	recAll := httptest.NewRecorder()
	e.ServeHTTP(recAll, reqAll)
	assert.Equal(t, http.StatusOK, recAll.Code)
	var allProds []ProductResponse
	json.Unmarshal(recAll.Body.Bytes(), &allProds)
	assert.Len(t, allProds, 1)

	// 4) PUT /products/:id (valid)
	updatePayload := map[string]interface{}{
		"name":        "Smartphone X",
		"description": "Upgraded phone",
		"price":       799.99,
		"category_id": catResp.ID,
	}
	updateBody, _ := json.Marshal(updatePayload)
	reqUpdate := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/products/%d", createdProd.ID), bytes.NewReader(updateBody))
	reqUpdate.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recUpdate := httptest.NewRecorder()
	e.ServeHTTP(recUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, recUpdate.Code)
	var updatedProd ProductResponse
	json.Unmarshal(recUpdate.Body.Bytes(), &updatedProd)
	assert.Equal(t, "Smartphone X", updatedProd.Name)

	// Negative: PUT /products/:id with invalid JSON
	reqUpdateBad := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/products/%d", createdProd.ID), bytes.NewReader([]byte(`{"price":"NaN"}`)))
	reqUpdateBad.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recUpdateBad := httptest.NewRecorder()
	e.ServeHTTP(recUpdateBad, reqUpdateBad)
	assert.Equal(t, http.StatusBadRequest, recUpdateBad.Code)

	// Negative: PUT /products/:id for non-existent ID
	updateBody2, _ := json.Marshal(updatePayload)
	reqUpdate2 := httptest.NewRequest(http.MethodPut, "/products/999", bytes.NewReader(updateBody2))
	reqUpdate2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recUpdate2 := httptest.NewRecorder()
	e.ServeHTTP(recUpdate2, reqUpdate2)
	assert.Equal(t, http.StatusNotFound, recUpdate2.Code)

	// 5) DELETE /products/:id (valid)
	reqDelete := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/products/%d", createdProd.ID), nil)
	recDelete := httptest.NewRecorder()
	e.ServeHTTP(recDelete, reqDelete)
	assert.Equal(t, http.StatusNoContent, recDelete.Code)

	// Negative: DELETE /products/:id with non-numeric ID
	reqDeleteBad := httptest.NewRequest(http.MethodDelete, "/products/abc", nil)
	recDeleteBad := httptest.NewRecorder()
	e.ServeHTTP(recDeleteBad, reqDeleteBad)
	assert.Equal(t, http.StatusNoContent, recDeleteBad.Code)
}

//--------------------------------------
// Category API tests
//--------------------------------------

func TestCategoryEndpoints_API(t *testing.T) {
	e, _ := setupAPIServer()

	// 1) POST /categories (valid)
	payload := map[string]interface{}{"name": "Books"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var catResp CategoryResponse
	json.Unmarshal(rec.Body.Bytes(), &catResp)
	assert.Equal(t, "Books", catResp.Name)

	// Negative: POST /categories with invalid JSON
	reqInv := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader([]byte(`{"name":123}`)))
	reqInv.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recInv := httptest.NewRecorder()
	e.ServeHTTP(recInv, reqInv)
	assert.Equal(t, http.StatusBadRequest, recInv.Code)

	// 2) GET /categories/:id (valid)
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/categories/%d", catResp.ID), nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	assert.Equal(t, http.StatusOK, recGet.Code)
	var fetchedCat CategoryResponse
	json.Unmarshal(recGet.Body.Bytes(), &fetchedCat)
	assert.Equal(t, catResp.ID, fetchedCat.ID)

	// Negative: GET /categories/:id with non-existent ID
	reqGetBad := httptest.NewRequest(http.MethodGet, "/categories/999", nil)
	recGetBad := httptest.NewRecorder()
	e.ServeHTTP(recGetBad, reqGetBad)
	assert.Equal(t, http.StatusNotFound, recGetBad.Code)

	// Negative: GET /categories/:id with non-numeric ID
	reqGetInv := httptest.NewRequest(http.MethodGet, "/categories/abc", nil)
	recGetInv := httptest.NewRecorder()
	e.ServeHTTP(recGetInv, reqGetInv)
	assert.Equal(t, http.StatusNotFound, recGetInv.Code)
}

//--------------------------------------
// Cart API tests
//--------------------------------------

func TestCartEndpoints_API(t *testing.T) {
	e, _ := setupAPIServer()

	// 1) POST /carts (valid)
	reqCart := httptest.NewRequest(http.MethodPost, "/carts", nil)
	recCart := httptest.NewRecorder()
	e.ServeHTTP(recCart, reqCart)
	assert.Equal(t, http.StatusCreated, recCart.Code)
	var cartResp CartResponse
	json.Unmarshal(recCart.Body.Bytes(), &cartResp)
	assert.NotZero(t, cartResp.ID)

	// Negative: POST /carts with unexpected payload (ignored by handler, still creates)
	reqCartBad := httptest.NewRequest(http.MethodPost, "/carts", bytes.NewReader([]byte(`{"foo":"bar"}`)))
	reqCartBad.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recCartBad := httptest.NewRecorder()
	e.ServeHTTP(recCartBad, reqCartBad)
	assert.Equal(t, http.StatusCreated, recCartBad.Code)

	// Create a product to add
	prodPayload := map[string]interface{}{
		"name":        "Headphones",
		"description": "Wireless headphones",
		"price":       199.99,
	}
	prodBody, _ := json.Marshal(prodPayload)
	reqProd := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(prodBody))
	reqProd.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recProd := httptest.NewRecorder()
	e.ServeHTTP(recProd, reqProd)
	var newProd ProductResponse
	json.Unmarshal(recProd.Body.Bytes(), &newProd)

	// 2) POST /carts/:id/products (valid)
	addPayload := map[string]interface{}{"product_id": newProd.ID}
	addBody, _ := json.Marshal(addPayload)
	reqAdd := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/carts/%d/products", cartResp.ID), bytes.NewReader(addBody))
	reqAdd.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAdd := httptest.NewRecorder()
	e.ServeHTTP(recAdd, reqAdd)
	assert.Equal(t, http.StatusOK, recAdd.Code)
	var updatedCart CartResponse
	json.Unmarshal(recAdd.Body.Bytes(), &updatedCart)
	assert.Len(t, updatedCart.Products, 1)

	// Negative: POST /carts/:id/products with invalid JSON
	reqAddInv := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/carts/%d/products", cartResp.ID), bytes.NewReader([]byte(`{"product_id":"NaN"}`)))
	reqAddInv.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAddInv := httptest.NewRecorder()
	e.ServeHTTP(recAddInv, reqAddInv)
	assert.Equal(t, http.StatusBadRequest, recAddInv.Code)

	// Negative: POST /carts/:id/products with non-numeric ID
	reqAddBadID := httptest.NewRequest(http.MethodPost, "/carts/abc/products", bytes.NewReader(addBody))
	reqAddBadID.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAddBadID := httptest.NewRecorder()
	e.ServeHTTP(recAddBadID, reqAddBadID)
	assert.Equal(t, http.StatusOK, recAddBadID.Code)
	var cartBadID CartResponse
	json.Unmarshal(recAddBadID.Body.Bytes(), &cartBadID)
	assert.Equal(t, uint(0), cartBadID.ID)

	// 3) GET /carts/:id (valid)
	reqGetCart := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/carts/%d", cartResp.ID), nil)
	recGetCart := httptest.NewRecorder()
	e.ServeHTTP(recGetCart, reqGetCart)
	assert.Equal(t, http.StatusOK, recGetCart.Code)
	var fetchedCart CartResponse
	json.Unmarshal(recGetCart.Body.Bytes(), &fetchedCart)
	assert.Len(t, fetchedCart.Products, 1)

	// Negative: GET /carts/:id with non-existent ID
	reqGetCartBad := httptest.NewRequest(http.MethodGet, "/carts/999", nil)
	recGetCartBad := httptest.NewRecorder()
	e.ServeHTTP(recGetCartBad, reqGetCartBad)
	assert.Equal(t, http.StatusOK, recGetCartBad.Code)
	var emptyCart CartResponse
	json.Unmarshal(recGetCartBad.Body.Bytes(), &emptyCart)
	assert.Equal(t, uint(0), emptyCart.ID)
	assert.Len(t, emptyCart.Products, 0)

	// Negative: GET /carts/:id with non-numeric ID
	reqGetCartInv := httptest.NewRequest(http.MethodGet, "/carts/abc", nil)
	recGetCartInv := httptest.NewRecorder()
	e.ServeHTTP(recGetCartInv, reqGetCartInv)
	assert.Equal(t, http.StatusOK, recGetCartInv.Code)
	var emptyCart2 CartResponse
	json.Unmarshal(recGetCartInv.Body.Bytes(), &emptyCart2)
	assert.Equal(t, uint(0), emptyCart2.ID)
	assert.Len(t, emptyCart2.Products, 0)
}

//--------------------------------------
// Miscellaneous negative tests
//--------------------------------------

// Invalid endpoint
func TestUnsupportedEndpoint(t *testing.T) {
	e, _ := setupAPIServer()
	req := httptest.NewRequest(http.MethodGet, "/checkout", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
