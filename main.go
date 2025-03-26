package main

import (
	"go_final/model"
	"net/http"
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

	// API Route for Login
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

// API Route for Change Password
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

	// Start server
	r.Run(":8080")
}




	

