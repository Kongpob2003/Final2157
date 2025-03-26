package main

import (
	"go_final/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectDatabase() (*gorm.DB, error) {
    dsn := "cp_65011212157:65011212157@csmsu@tcp(202.28.34.197:3306)/cp_65011212157?parseTime=true"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}

func main() {
	r := gin.Default()

	// Connect to database
	db, err := connectDatabase()
	if err != nil {
		panic("Failed to connect to database!")
	}

	//login
r.POST("/login", func(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Binding input JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var customer model.Customer
	// Find customer by email
	if err := db.Where("email = ?", input.Email).First(&customer).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check if password is already hashed or not
	var passwordMatch bool
	if len(customer.Password) < 60 { // This assumes that bcrypt hashes are always 60 characters long
		// If password is not hashed (i.e. it's in plain text), compare directly
		if customer.Password == input.Password {
			passwordMatch = true
		}
	} else {
		// If password is hashed, compare with bcrypt
		err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(input.Password))
		if err == nil {
			passwordMatch = true
		}
	}

	// If password matches, proceed
	if passwordMatch {
		// If password was plain, hash it and save to database
		if len(customer.Password) < 60 {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}
			customer.Password = string(hashedPassword)
			// Update the customer's password with the hashed password
			if err := db.Save(&customer).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
				return
			}
		}

		// Return customer data in response
		c.JSON(http.StatusOK, gin.H{
			"CustomerID":  customer.CustomerID,
			"FirstName":   customer.FirstName,
			"LastName":    customer.LastName,
			"Email":       customer.Email,
			"PhoneNumber": customer.PhoneNumber,
			"Address":     customer.Address,
			"CreatedAt":   customer.CreatedAt.Format(time.RFC3339),
			"UpdatedAt":   customer.UpdatedAt.Format(time.RFC3339),
		})
		
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
	}
})

	//Change Password
r.POST("/changepassword", func(c *gin.Context) {
	var input struct {
		Email       string `json:"email" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Find customer by email
	var customer model.Customer
	if err := db.Where("email = ?", input.Email).First(&customer).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
		return
	}

	// Verify old password
	var passwordMatch bool
	if len(customer.Password) < 60 { // Plain text password
		if customer.Password == input.OldPassword {
			passwordMatch = true
		}
	} else { // Hashed password
		err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(input.OldPassword))
		if err == nil {
			passwordMatch = true
		}
	}

	if !passwordMatch {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	// Hash the new password
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update the password in the database
	customer.Password = string(hashedNewPassword)
	if err := db.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
})

	// API Route for Adding Product to Cart
	r.POST("/addtocart", func(c *gin.Context) {
		var input struct {
			Email       string  `json:"email" binding:"required"`
			CartName    string  `json:"cart_name" binding:"required"`
			Description string  `json:"description"`           // Optional: search by product description
			MinPrice    float64 `json:"min_price"`             // Optional: minimum price
			MaxPrice    float64 `json:"max_price"`             // Optional: maximum price
			ProductID   int     `json:"product_id" binding:"required"`
			Quantity    int     `json:"quantity" binding:"required,gte=1"`
		}

		// Bind JSON input
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Find customer by email
		var customer model.Customer
		if err := db.Where("email = ?", input.Email).First(&customer).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
			return
		}

		// Search for the product
		var product model.Product
		query := db.Where("product_id = ?", input.ProductID)
		if input.Description != "" {
			query = query.Where("description LIKE ?", "%"+input.Description+"%")
		}
		if input.MinPrice > 0 {
			priceMin := strconv.FormatFloat(input.MinPrice, 'f', -1, 64)
			query = query.Where("price >= ?", priceMin)
		}
		if input.MaxPrice > 0 {
			priceMax := strconv.FormatFloat(input.MaxPrice, 'f', -1, 64)
			query = query.Where("price <= ?", priceMax)
		}
		if err := query.First(&product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// Check stock availability
		if product.StockQuantity < input.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
			return
		}

		// Find or create the cart
		var cart model.Cart
		if err := db.Where("customer_id = ? AND cart_name = ?", customer.CustomerID, input.CartName).First(&cart).Error; err != nil {
			// Cart doesn't exist, create a new one
			cart = model.Cart{
				CustomerID: customer.CustomerID,
				CartName:   input.CartName,
			}
			if err := db.Create(&cart).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
				return
			}
		}

		// Check if the product already exists in the cart
		var cartItem model.CartItem
		if err := db.Where("cart_id = ? AND product_id = ?", cart.CartID, product.ProductID).First(&cartItem).Error; err == nil {
			// Product exists, update quantity
			cartItem.Quantity += input.Quantity
			if err := db.Save(&cartItem).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
				return
			}
		} else {
			// Product doesn't exist, add new cart item
			cartItem = model.CartItem{
				CartID:    cart.CartID,
				ProductID: product.ProductID,
				Quantity:  input.Quantity,
			}
			if err := db.Create(&cartItem).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
				return
			}
		}

		// Success response
		c.JSON(http.StatusOK, gin.H{
			"message":    "Product added to cart successfully",
			"cart_id":    cart.CartID,
			"cart_name":  cart.CartName,
			"product_id": product.ProductID,
			"quantity":   cartItem.Quantity,
		})
	})

	// API Route for Search Products
	r.GET("/search-products", func(c *gin.Context) {
		var input model.SearchProductInput

		// Bind query parameters
		if err := c.ShouldBindQuery(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Build the query
		var products []model.Product
		query := db.Model(&model.Product{})
		if input.Description != "" {
			query = query.Where("description LIKE ?", "%"+input.Description+"%")
		}
		if input.MinPrice > 0 {
			priceMin := strconv.FormatFloat(input.MinPrice, 'f', -1, 64)
			query = query.Where("price >= ?", priceMin)
		}
		if input.MaxPrice > 0 {
			priceMax := strconv.FormatFloat(input.MaxPrice, 'f', -1, 64)
			query = query.Where("price <= ?", priceMax)
		}

		// Execute the query
		if err := query.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
			return
		}

		// Format the response
		response := make([]gin.H, len(products))
		for i, product := range products {
			response[i] = gin.H{
				"product_id":     product.ProductID,
				"product_name":   product.ProductName,
				"description":    product.Description,
				"price":          product.Price,
				"stock_quantity": product.StockQuantity,
				"created_at":     product.CreatedAt.Format(time.RFC3339),
				"updated_at":     product.UpdatedAt.Format(time.RFC3339),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Products retrieved successfully",
			"data":    response,
		})
	})
	// Start server
	r.Run(":8080")
}




	

