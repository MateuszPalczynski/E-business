// tests/pages/product_page.go
package pages

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tebeka/selenium"
)

type ProductPage struct {
	Driver         selenium.WebDriver
	QuantityInput  string
	AddToCartBtn   string
	ProductPrice   string
	ProductTitle   string
	QuantityError  string
}

func NewProductPage(driver selenium.WebDriver) *ProductPage {
	return &ProductPage{
		Driver:         driver,
		QuantityInput:  "quantity-input",
		AddToCartBtn:   "add-to-cart-button",
		ProductPrice:   "product-price",
		ProductTitle:   "product-title",
		QuantityError:  "quantity-error",
	}
}

func (p *ProductPage) VerifyProductDetails(title string, price float64) error {
	// Verify product title
	actualTitle, err := p.Driver.FindElement(selenium.ByCSSSelector, p.ProductTitle)
	if err != nil {
		return fmt.Errorf("error finding product title: %v", err)
	}
	
	txt, _ := actualTitle.Text()
	if txt != title {
		return fmt.Errorf("expected title '%s', got '%s'", title, txt)
	}

	// Verify product price
	priceElement, err := p.Driver.FindElement(selenium.ByCSSSelector, p.ProductPrice)
	if err != nil {
		return fmt.Errorf("error finding product price: %v", err)
	}

	priceText, _ := priceElement.Text()
	currentPrice, err := strconv.ParseFloat(priceText, 64)
	if err != nil {
		return fmt.Errorf("error parsing price: %v", err)
	}

	if currentPrice != price {
		return fmt.Errorf("expected price %.2f, got %.2f", price, currentPrice)
	}

	return nil
}

func (p *ProductPage) SetQuantity(quantity int) error {
	input, err := p.Driver.FindElement(selenium.ByCSSSelector, p.QuantityInput)
	if err != nil {
		return fmt.Errorf("error finding quantity input: %v", err)
	}

	if err := input.Clear(); err != nil {
		return fmt.Errorf("error clearing quantity input: %v", err)
	}

	if err := input.SendKeys(strconv.Itoa(quantity)); err != nil {
		return fmt.Errorf("error setting quantity: %v", err)
	}

	return nil
}

func (p *ProductPage) AddToCart() error {
	btn, err := p.Driver.FindElement(selenium.ByCSSSelector, p.AddToCartBtn)
	if err != nil {
		return fmt.Errorf("error finding add to cart button: %v", err)
	}

	if err := btn.Click(); err != nil {
		return fmt.Errorf("error clicking add to cart button: %v", err)
	}

	// Wait for animation/redirect
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (p *ProductPage) VerifyQuantity(expected int) error {
	input, err := p.Driver.FindElement(selenium.ByCSSSelector, p.QuantityInput)
	if err != nil {
		return fmt.Errorf("error finding quantity input: %v", err)
	}

	value, err := input.GetAttribute("value")
	if err != nil {
		return fmt.Errorf("error getting quantity value: %v", err)
	}

	actual, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("error converting quantity to integer: %v", err)
	}

	if actual != expected {
		return fmt.Errorf("expected quantity %d, got %d", expected, actual)
	}

	return nil
}

func (p *ProductPage) VerifyQuantityError(expectedMessage string) error {
	errorElement, err := p.Driver.FindElement(selenium.ByCSSSelector, p.QuantityError)
	if err != nil {
		return fmt.Errorf("error finding quantity error element: %v", err)
	}

	actualMessage, _ := errorElement.Text()
	if actualMessage != expectedMessage {
		return fmt.Errorf("expected error message '%s', got '%s'", expectedMessage, actualMessage)
	}

	return nil
}