package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Helper to initialize Echo instance with isolated in-memory SQLite and routes
func setupTestServer() (*echo.Echo, *gorm.DB) {
	// In-memory SQLite DB (isolated per connection)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&Product{}, &Cart{}, &Category{})

	e := echo.New()
	e.Use(DBMiddleware(db))

	// Raw test endpoint
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

// Structs for decoding JSON responses
type ProductResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CategoryID  uint      `json:"category_id"`
	Category    Category  `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryResponse struct {
	ID        uint              `json:"id"`
	Name      string            `json:"name"`
	Products  []ProductResponse `json:"products"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type CartResponse struct {
	ID        uint              `json:"id"`
	Products  []ProductResponse `json:"products"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

//--------------------------------------
// Begin tests
//--------------------------------------

func TestRawEndpoint(t *testing.T) {
	e, _ := setupTestServer()
	// Prepare request
	payload := `{"foo":"bar"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(payload)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)        // 1
	assert.Equal(t, payload, rec.Body.String())     // 2
	assert.Contains(t, rec.Body.String(), `"foo":"bar"`) // 3
}

//--------------------------------------
// Product tests
//--------------------------------------

func TestCreateGetUpdateDeleteProduct(t *testing.T) {
	e, _ := setupTestServer()

	// ---- Create Category for attaching to product ----
	catPayload := map[string]interface{}{"name": "Electronics"}
	catBody, _ := json.Marshal(catPayload)
	reqCat := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(catBody))
	reqCat.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recCat := httptest.NewRecorder()
	e.ServeHTTP(recCat, reqCat)
	assert.Equal(t, http.StatusCreated, recCat.Code) // 4

	var catResp CategoryResponse
	err := json.Unmarshal(recCat.Body.Bytes(), &catResp)
	assert.NoError(t, err)                // 5
	assert.Equal(t, uint(1), catResp.ID)  // 6
	assert.Equal(t, "Electronics", catResp.Name) // 7

	// ---- Create Product ----
	prodPayload := map[string]interface{}{
		"name":         "Laptop",
		"description":  "Gaming laptop",
		"price":        1299.99,
		"category_id":  catResp.ID,
	}
	prodBody, _ := json.Marshal(prodPayload)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(prodBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code) // 8

	var createdProd ProductResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createdProd)
	assert.NoError(t, err)                       // 9
	assert.Equal(t, "Laptop", createdProd.Name)   // 10
	assert.Equal(t, "Gaming laptop", createdProd.Description) // 11
	assert.InEpsilon(t, 1299.99, createdProd.Price, 0.001)   // 12
	assert.Equal(t, catResp.ID, createdProd.CategoryID)     // 13
	assert.NotZero(t, createdProd.ID)                       // 14
	assert.WithinDuration(t, time.Now(), createdProd.CreatedAt, time.Second*5) // 15
	assert.WithinDuration(t, time.Now(), createdProd.UpdatedAt, time.Second*5) // 16

	// ---- Get Product by ID ----
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%d", createdProd.ID), nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	assert.Equal(t, http.StatusOK, recGet.Code) // 17

	var fetchedProd ProductResponse
	err = json.Unmarshal(recGet.Body.Bytes(), &fetchedProd)
	assert.NoError(t, err)                    // 18
	assert.Equal(t, createdProd.ID, fetchedProd.ID)   // 19
	assert.Equal(t, createdProd.Name, fetchedProd.Name) // 20
	assert.Equal(t, fetchedProd.Category.Name, "Electronics") // 21

	// ---- Update Product ----
	updatePayload := map[string]interface{}{
		"name":         "Laptop Pro",
		"description":  "High-end gaming laptop",
		"price":        1499.49,
		"category_id":  catResp.ID,
	}
	updateBody, _ := json.Marshal(updatePayload)
	reqUpdate := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/products/%d", createdProd.ID), bytes.NewReader(updateBody))
	reqUpdate.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recUpdate := httptest.NewRecorder()
	e.ServeHTTP(recUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, recUpdate.Code) // 22

	var updatedProd ProductResponse
	err = json.Unmarshal(recUpdate.Body.Bytes(), &updatedProd)
	assert.NoError(t, err)                                // 23
	assert.Equal(t, "Laptop Pro", updatedProd.Name)        // 24
	assert.Equal(t, "High-end gaming laptop", updatedProd.Description) // 25
	assert.InEpsilon(t, 1499.49, updatedProd.Price, 0.001) // 26
	assert.True(t, updatedProd.UpdatedAt.After(updatedProd.CreatedAt)) // 27

	// ---- Get All Products (should be 1) ----
	reqAll := httptest.NewRequest(http.MethodGet, "/products", nil)
	recAll := httptest.NewRecorder()
	e.ServeHTTP(recAll, reqAll)
	assert.Equal(t, http.StatusOK, recAll.Code) // 28

	var allProds []ProductResponse
	err = json.Unmarshal(recAll.Body.Bytes(), &allProds)
	assert.NoError(t, err)   // 29
	assert.Len(t, allProds, 1) // 30

	// ---- Delete Product ----
	reqDelete := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/products/%d", createdProd.ID), nil)
	recDelete := httptest.NewRecorder()
	e.ServeHTTP(recDelete, reqDelete)
	assert.Equal(t, http.StatusNoContent, recDelete.Code) // 31

	// ---- Verify Deletion: Get /products/:id returns empty product ----
	reqGet2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%d", createdProd.ID), nil)
	recGet2 := httptest.NewRecorder()
	e.ServeHTTP(recGet2, reqGet2)
	assert.Equal(t, http.StatusOK, recGet2.Code) // 32

	var emptyProd ProductResponse
	err = json.Unmarshal(recGet2.Body.Bytes(), &emptyProd)
	assert.NoError(t, err)        // 33
	assert.Equal(t, uint(0), emptyProd.ID) // 34
}

// Test updating non-existent product returns 404
func TestUpdateNonExistentProduct(t *testing.T) {
	e, _ := setupTestServer()
	updatePayload := map[string]interface{}{
		"name":        "Ghost",
		"description": "Doesn't exist",
		"price":       0.0,
	}
	updateBody, _ := json.Marshal(updatePayload)
	req := httptest.NewRequest(http.MethodPut, "/products/999", bytes.NewReader(updateBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code) // 35
}

//--------------------------------------
// Category tests
//--------------------------------------

func TestCreateGetCategory(t *testing.T) {
	e, _ := setupTestServer()

	// ---- Create Category ----
	payload := map[string]interface{}{"name": "Books"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code) // 36

	var catResp CategoryResponse
	err := json.Unmarshal(rec.Body.Bytes(), &catResp)
	assert.NoError(t, err)       // 37
	assert.Equal(t, "Books", catResp.Name) // 38
	assert.NotZero(t, catResp.ID) // 39

	// ---- Get Category by ID ----
	reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/categories/%d", catResp.ID), nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)

	assert.Equal(t, http.StatusOK, recGet.Code) // 40

	var fetchedCat CategoryResponse
	err = json.Unmarshal(recGet.Body.Bytes(), &fetchedCat)
	assert.NoError(t, err)               // 41
	assert.Equal(t, catResp.ID, fetchedCat.ID) // 42
	assert.Equal(t, "Books", fetchedCat.Name)  // 43
	assert.Len(t, fetchedCat.Products, 0)      // 44

	// ---- Get non-existent Category ----
	reqNotFound := httptest.NewRequest(http.MethodGet, "/categories/999", nil)
	recNotFound := httptest.NewRecorder()
	e.ServeHTTP(recNotFound, reqNotFound)
	assert.Equal(t, http.StatusNotFound, recNotFound.Code) // 45
}

//--------------------------------------
// Cart tests
//--------------------------------------

func TestCartLifecycle(t *testing.T) {
	e, _ := setupTestServer()

	// ---- Create Cart ----
	reqCart := httptest.NewRequest(http.MethodPost, "/carts", nil)
	recCart := httptest.NewRecorder()
	e.ServeHTTP(recCart, reqCart)
	assert.Equal(t, http.StatusCreated, recCart.Code) // 46

	var cartResp CartResponse
	err := json.Unmarshal(recCart.Body.Bytes(), &cartResp)
	assert.NoError(t, err)      // 47
	assert.NotZero(t, cartResp.ID) // 48
	assert.Len(t, cartResp.Products, 0) // 49

	// ---- Create a Product to add into cart ----
	prodPayload := map[string]interface{}{
		"name":         "Tablet",
		"description":  "Android tablet",
		"price":        299.99,
		"category_id":  0,
	}
	prodBody, _ := json.Marshal(prodPayload)
	reqProd := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(prodBody))
	reqProd.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recProd := httptest.NewRecorder()
	e.ServeHTTP(recProd, reqProd)
	assert.Equal(t, http.StatusCreated, recProd.Code) // 50

	var newProd ProductResponse
	err = json.Unmarshal(recProd.Body.Bytes(), &newProd)
	assert.NoError(t, err) // 51

	// ---- Add Product to Cart ----
	addPayload := map[string]interface{}{"product_id": newProd.ID}
	addBody, _ := json.Marshal(addPayload)
	reqAdd := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/carts/%d/products", cartResp.ID), bytes.NewReader(addBody))
	reqAdd.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAdd := httptest.NewRecorder()
	e.ServeHTTP(recAdd, reqAdd)
	assert.Equal(t, http.StatusOK, recAdd.Code) // 52

	var updatedCart CartResponse
	err = json.Unmarshal(recAdd.Body.Bytes(), &updatedCart)
	assert.NoError(t, err)                 // 53
	assert.Equal(t, cartResp.ID, updatedCart.ID) // 54
	assert.Len(t, updatedCart.Products, 1)  // 55
	assert.Equal(t, "Tablet", updatedCart.Products[0].Name) // 56

	// ---- Get Cart with Product ----
	reqGetCart := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/carts/%d", cartResp.ID), nil)
	recGetCart := httptest.NewRecorder()
	e.ServeHTTP(recGetCart, reqGetCart)
	assert.Equal(t, http.StatusOK, recGetCart.Code) // 57

	var fetchedCart CartResponse
	err = json.Unmarshal(recGetCart.Body.Bytes(), &fetchedCart)
	assert.NoError(t, err)                // 58
	assert.Len(t, fetchedCart.Products, 1) // 59
	assert.Equal(t, newProd.ID, fetchedCart.Products[0].ID) // 60
}

//--------------------------------------
// Miscellaneous negative tests
//--------------------------------------

// Test getAllProducts returns empty list when none created
func TestGetAllProductsEmpty(t *testing.T) {
	e, _ := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code) // 61
	var prods []ProductResponse
	err := json.Unmarshal(rec.Body.Bytes(), &prods)
	assert.NoError(t, err)     // 62
	assert.Len(t, prods, 0)    // 63
}

func TestGetCartEmptyID(t *testing.T) {
	e, _ := setupTestServer()
	// Attempt to GET non-existent cart
	req := httptest.NewRequest(http.MethodGet, "/carts/999", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code) // 64
	var emptyCart CartResponse
	err := json.Unmarshal(rec.Body.Bytes(), &emptyCart)
	assert.NoError(t, err)            // 65
	assert.Equal(t, uint(0), emptyCart.ID) // 66
	assert.Len(t, emptyCart.Products, 0)   // 67
}

func TestAddProductToNonexistentCart(t *testing.T) {
	e, _ := setupTestServer()
	// Create a product
	prodPayload := map[string]interface{}{
		"name":        "Mouse",
		"description": "Wireless mouse",
		"price":       49.99,
	}
	prodBody, _ := json.Marshal(prodPayload)
	reqProd := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(prodBody))
	reqProd.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recProd := httptest.NewRecorder()
	e.ServeHTTP(recProd, reqProd)
	assert.Equal(t, http.StatusCreated, recProd.Code) // 68

	var p ProductResponse
	_ = json.Unmarshal(recProd.Body.Bytes(), &p)

	// Attempt to add product to cart ID 999 (non-existent)
	addPayload := map[string]interface{}{"product_id": p.ID}
	addBody, _ := json.Marshal(addPayload)
	reqAdd := httptest.NewRequest(http.MethodPost, "/carts/999/products", bytes.NewReader(addBody))
	reqAdd.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAdd := httptest.NewRecorder()
	e.ServeHTTP(recAdd, reqAdd)
	assert.Equal(t, http.StatusOK, recAdd.Code) // 69

	var cartResp CartResponse
	err := json.Unmarshal(recAdd.Body.Bytes(), &cartResp)
	assert.NoError(t, err)                         // 70
	assert.Equal(t, uint(0), cartResp.ID)          // 71
	assert.Len(t, cartResp.Products, 1)            // 72
}
