import unittest

from selenium import webdriver
from selenium.common.exceptions import TimeoutException
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait


class SauceDemoTests(unittest.TestCase):
    def setUp(self):
        # Initialize the WebDriver (Chrome)
        self.driver = webdriver.Chrome()
        self.driver.implicitly_wait(5)
        self.driver.maximize_window()
        # Open the SauceDemo login page
        self.driver.get("https://www.saucedemo.com/")

    def tearDown(self):
        # Quit the WebDriver
        self.driver.quit()

    def login(self, username, password):
        """Helper for logging in."""
        driver = self.driver
        driver.find_element(By.ID, "user-name").send_keys(username)
        driver.find_element(By.ID, "password").send_keys(password)
        driver.find_element(By.ID, "login-button").click()

    def navigate_to_cart(self):
        """Helper: click cart link and wait for cart page to load."""
        driver = self.driver
        driver.find_element(By.CLASS_NAME, "shopping_cart_link").click()
        # Wait until the cart page title is visible
        WebDriverWait(driver, 5).until(
            EC.text_to_be_present_in_element((By.CLASS_NAME, "title"), "Your Cart")
        )

    def test_valid_login(self):
        """Test successful login with valid credentials."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Verify landing on the inventory page: header text is 'Products'
        title = driver.find_element(By.CLASS_NAME, "title").text
        self.assertEqual(title, "Products")

        # Verify URL contains 'inventory'
        self.assertIn("inventory", driver.current_url)

        # Verify inventory container is present
        inv_container = driver.find_element(By.ID, "inventory_container")
        self.assertTrue(inv_container.is_displayed())

        # Verify the number of products listed (should be 6)
        items = driver.find_elements(By.CLASS_NAME, "inventory_item")
        self.assertEqual(len(items), 6)

        # Verify filter dropdown is present
        sort_select = driver.find_element(By.CLASS_NAME, "product_sort_container")
        self.assertTrue(sort_select.is_displayed())

    def test_invalid_login_wrong_password(self):
        """Test login with invalid password."""
        self.login("standard_user", "wrong_password")
        error = self.driver.find_element(By.CSS_SELECTOR, "[data-test='error']").text
        self.assertIn("Epic sadface", error)
        self.assertIn("do not match any user", error)

    def test_invalid_login_wrong_username(self):
        """Test login with invalid username."""
        self.login("invalid_user", "secret_sauce")
        error = self.driver.find_element(By.CSS_SELECTOR, "[data-test='error']").text
        self.assertIn("Epic sadface", error)
        self.assertIn("do not match any user", error)

    def test_invalid_login_empty_username(self):
        """Test login with empty username field."""
        self.login("", "secret_sauce")
        error = self.driver.find_element(By.CSS_SELECTOR, "[data-test='error']").text
        self.assertIn("Epic sadface", error)
        self.assertIn("Username is required", error)

    def test_invalid_login_empty_password(self):
        """Test login with empty password field."""
        self.login("standard_user", "")
        error = self.driver.find_element(By.CSS_SELECTOR, "[data-test='error']").text
        self.assertIn("Epic sadface", error)
        self.assertIn("Password is required", error)

    def test_login_locked_out_user(self):
        """Test login with a locked out user."""
        self.login("locked_out_user", "secret_sauce")
        error = self.driver.find_element(By.CSS_SELECTOR, "[data-test='error']").text
        self.assertIn("Epic sadface", error)
        self.assertIn("locked out", error)

    def test_verify_elements_after_login(self):
        """Verify presence of key elements after successful login."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Check page header
        title = driver.find_element(By.CLASS_NAME, "title").text
        self.assertEqual(title, "Products")

        # Check inventory items count
        inventory_items = driver.find_elements(By.CLASS_NAME, "inventory_item")
        self.assertEqual(len(inventory_items), 6)

        # Verify each product name is not empty
        names = driver.find_elements(By.CLASS_NAME, "inventory_item_name")
        for name in names:
            self.assertNotEqual(name.text, "")

        # Verify each product price has correct format
        prices = driver.find_elements(By.CLASS_NAME, "inventory_item_price")
        for price in prices:
            self.assertRegex(price.text, r"^\$\d+\.\d{2}$")

        # Verify filter dropdown exists
        self.assertTrue(driver.find_element(By.CLASS_NAME, "product_sort_container").is_displayed())

    def test_logout(self):
        """Test the logout functionality."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Open the side menu
        driver.find_element(By.ID, "react-burger-menu-btn").click()
        # Wait for logout link to be visible
        try:
            logout_link = WebDriverWait(driver, 5).until(
                EC.visibility_of_element_located((By.ID, "logout_sidebar_link"))
            )
            logout_link.click()
        except TimeoutException:
            self.fail("Logout link was not visible in time.")

        # After logout, verify that login button is present again
        login_btn = driver.find_element(By.ID, "login-button")
        self.assertTrue(login_btn.is_displayed())

    def test_add_single_product_to_cart(self):
        """Test adding a single product to the cart."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add the Sauce Labs Backpack to the cart
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()

        # Navigate to cart and verify exactly one item is present
        self.navigate_to_cart()
        cart_items = driver.find_elements(By.CLASS_NAME, "inventory_item_name")
        self.assertEqual(len(cart_items), 1)
        self.assertEqual(cart_items[0].text, "Sauce Labs Backpack")

    def test_add_multiple_products_to_cart(self):
        """Test adding multiple products to the cart."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add three products
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()
        driver.find_element(By.ID, "add-to-cart-sauce-labs-bolt-t-shirt").click()
        driver.find_element(By.ID, "add-to-cart-sauce-labs-bike-light").click()

        # Navigate to cart
        self.navigate_to_cart()
        cart_items = driver.find_elements(By.CLASS_NAME, "inventory_item_name")

        # Verify three items are present
        self.assertEqual(len(cart_items), 3)

    def test_remove_products_from_cart(self):
        """Test removing products from the cart."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add two products
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()
        driver.find_element(By.ID, "add-to-cart-sauce-labs-bike-light").click()

        # Navigate to cart
        self.navigate_to_cart()
        # Remove one product
        remove_btn = WebDriverWait(driver, 5).until(
            EC.visibility_of_element_located((By.ID, "remove-sauce-labs-backpack"))
        )
        remove_btn.click()
        remaining_items = [item.text for item in driver.find_elements(By.CLASS_NAME, "inventory_item_name")]
        self.assertListEqual(remaining_items, ["Sauce Labs Bike Light"])

        # Remove second product
        remove_second = WebDriverWait(driver, 5).until(
            EC.visibility_of_element_located((By.ID, "remove-sauce-labs-bike-light"))
        )
        remove_second.click()
        # Verify cart is empty
        cart_items = driver.find_elements(By.CLASS_NAME, "inventory_item_name")
        self.assertEqual(len(cart_items), 0)

    def test_continue_shopping_button(self):
        """Test the 'Continue Shopping' button in the cart."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add an item and go to cart
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()
        self.navigate_to_cart()

        # Click continue shopping
        continue_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "continue-shopping"))
        )
        continue_btn.click()

        # Verify we are back on the inventory page
        title = driver.find_element(By.CLASS_NAME, "title").text
        self.assertEqual(title, "Products")
        self.assertIn("inventory", driver.current_url)

    def test_checkout_complete_process(self):
        """Test complete checkout flow with valid data."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add item to cart and proceed to checkout
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()
        self.navigate_to_cart()
        checkout_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "checkout"))
        )
        checkout_btn.click()

        # Fill in checkout information
        driver.find_element(By.ID, "first-name").send_keys("John")
        driver.find_element(By.ID, "last-name").send_keys("Doe")
        driver.find_element(By.ID, "postal-code").send_keys("12345")
        driver.find_element(By.ID, "continue").click()

        # Finish checkout
        finish_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "finish"))
        )
        finish_btn.click()

        # Verify checkout complete page
        complete_header = driver.find_element(By.CLASS_NAME, "complete-header").text
        self.assertEqual(complete_header, "Thank you for your order!")
        complete_text = driver.find_element(By.CLASS_NAME, "complete-text").text
        self.assertIn("Your order has been dispatched", complete_text)

    def test_checkout_missing_info(self):
        """Test checkout with missing information to trigger an error."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add item and begin checkout
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()
        self.navigate_to_cart()
        checkout_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "checkout"))
        )
        checkout_btn.click()

        # Leave out last name
        driver.find_element(By.ID, "first-name").send_keys("John")
        driver.find_element(By.ID, "postal-code").send_keys("12345")
        driver.find_element(By.ID, "continue").click()

        # Verify error message for missing last name
        error = driver.find_element(By.CSS_SELECTOR, "[data-test='error']").text
        self.assertIn("Last Name is required", error)

    def test_checkout_cancel(self):
        """Test cancelling from the checkout information page."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add item and begin checkout
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()
        self.navigate_to_cart()
        checkout_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "checkout"))
        )
        checkout_btn.click()

        # Cancel on the information page
        cancel_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "cancel"))
        )
        cancel_btn.click()

        # Verify we are back at the cart page
        self.assertIn("cart", driver.current_url)
        page_title = driver.find_element(By.CLASS_NAME, "title").text
        self.assertEqual(page_title, "Your Cart")

    def test_filter_sort_name_az(self):
        """Test sorting products by name A to Z."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Sort by Name (A to Z)
        select = driver.find_element(By.CLASS_NAME, "product_sort_container")
        select.click()
        select.find_element(By.XPATH, "//option[@value='az']").click()

        # Verify product names are in ascending order
        names = [elem.text for elem in driver.find_elements(By.CLASS_NAME, "inventory_item_name")]
        self.assertEqual(names, sorted(names))

    def test_filter_sort_name_za(self):
        """Test sorting products by name Z to A."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Sort by Name (Z to A)
        select = driver.find_element(By.CLASS_NAME, "product_sort_container")
        select.click()
        select.find_element(By.XPATH, "//option[@value='za']").click()

        names = [elem.text for elem in driver.find_elements(By.CLASS_NAME, "inventory_item_name")]
        self.assertEqual(names, sorted(names, reverse=True))

    def test_filter_sort_price_low_to_high(self):
        """Test sorting products by price low to high."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Sort by Price (low to high)
        select = driver.find_element(By.CLASS_NAME, "product_sort_container")
        select.click()
        select.find_element(By.XPATH, "//option[@value='lohi']").click()

        prices = [float(elem.text.replace('$', '')) for elem in driver.find_elements(By.CLASS_NAME, "inventory_item_price")]
        self.assertEqual(prices, sorted(prices))

    def test_filter_sort_price_high_to_low(self):
        """Test sorting products by price high to low."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Sort by Price (high to low)
        select = driver.find_element(By.CLASS_NAME, "product_sort_container")
        select.click()
        select.find_element(By.XPATH, "//option[@value='hilo']").click()

        prices = [float(elem.text.replace('$', '')) for elem in driver.find_elements(By.CLASS_NAME, "inventory_item_price")]
        self.assertEqual(prices, sorted(prices, reverse=True))

    def test_order_summary_total_validation(self):
        """Test that the order summary item total is correct."""
        driver = self.driver
        self.login("standard_user", "secret_sauce")

        # Add two known products
        driver.find_element(By.ID, "add-to-cart-sauce-labs-backpack").click()  # $29.99
        driver.find_element(By.ID, "add-to-cart-sauce-labs-bike-light").click()  # $9.99

        # Go to cart and checkout
        self.navigate_to_cart()
        checkout_btn = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.ID, "checkout"))
        )
        checkout_btn.click()

        driver.find_element(By.ID, "first-name").send_keys("Anna")
        driver.find_element(By.ID, "last-name").send_keys("Smith")
        driver.find_element(By.ID, "postal-code").send_keys("54321")
        driver.find_element(By.ID, "continue").click()

        # Calculate expected total from listed prices
        prices = [float(elem.text.replace('$', '')) for elem in driver.find_elements(By.CLASS_NAME, "inventory_item_price")]
        expected_sum = sum(prices)

        # Get the item total from the summary
        item_total_text = driver.find_element(By.CLASS_NAME, "summary_subtotal_label").text
        total_value = float(item_total_text.replace("Item total: $", ""))
        self.assertAlmostEqual(expected_sum, total_value, places=2)

if __name__ == "__main__":
    unittest.main()
