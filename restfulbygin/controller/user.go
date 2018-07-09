package controller

import (
	"net/http"
	"time"

	"github.com/fourth04/initialize/restfulbygin/model"
	"github.com/fourth04/initialize/restfulbygin/utils"
	"github.com/gin-gonic/gin"
)

func PostUser(c *gin.Context) {
	var user model.User
	c.Bind(&user)

	if user.Username != "" && user.Password != "" && user.RoleName != "" && user.RateFormatted != "" {
		var oldUser model.User
		model.DB.Where("username = ?", user.Username).First(&oldUser)
		if oldUser.ID == 0 {
			if user.RoleName == "admin" || user.RoleName == "user" {
				salt := utils.RandomString(10)
				password, err := utils.Encrypt(user.Password, salt)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
				user.Salt = salt
				user.Password = password
				user.CreatedAt = time.Now()
				user.UpdatedAt = time.Now()
				// INSERT INTO "users" (name) VALUES (user.Name);
				model.DB.Create(&user)
				// Display error
				c.JSON(http.StatusOK, gin.H{"success": user.Desentitize()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "role_name field must be admin or user"})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User exists"})
		}
	} else {
		// Display error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username or password or role_name or rate_formatted field is empty"})
	}

	// curl -i -X POST -H "Content-Type: application/json" -d "{ \"username\": \"Thea\", \"password\": \"Queen\" , \"role_name\": \"admin\"}" http://localhost:8080/api/v1/users
}

func GetUsers(c *gin.Context) {
	var users []model.User
	// SELECT * FROM users
	model.DB.Find(&users)

	usersDesensitized := make([]interface{}, len(users))

	for i, user := range users {
		usersDesensitized[i] = user.Desentitize()
	}

	// Display JSON result
	c.JSON(http.StatusOK, gin.H{"success": usersDesensitized})

	// curl -i http://localhost:8080/api/v1/users
}

func GetUser(c *gin.Context) {

	id := c.Params.ByName("id")
	var user model.User
	// SELECT * FROM users WHERE id = 1;
	model.DB.First(&user, id)

	if user.ID != 0 {
		// Display JSON result
		c.JSON(http.StatusOK, gin.H{"success": user.Desentitize()})
	} else {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
	}

	// curl -i http://localhost:8080/api/v1/users/1
}

func UpdateUser(c *gin.Context) {

	// Get id user
	id := c.Params.ByName("id")
	var user model.User
	// SELECT * FROM users WHERE id = 1;
	model.DB.First(&user, id)
	if user.ID != 0 {
		var newUser model.User
		c.Bind(&newUser)
		if newUser.Username != "" && newUser.Password != "" && newUser.RoleName != "" && newUser.RateFormatted != "" {
			if newUser.RoleName == "admin" || newUser.RoleName == "newUser" {
				newUser.ID = user.ID
				newUser.CreatedAt = user.CreatedAt

				salt := utils.RandomString(10)
				password, err := utils.Encrypt(newUser.Password, salt)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
				newUser.Salt = salt
				newUser.Password = password
				newUser.UpdatedAt = time.Now()

				// UPDATE users SET firstname='newUser.Username', lastname='newUser.Password' WHERE id = user.ID;
				model.DB.Save(&newUser)
				err = model.DB.Save(&newUser).Error
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				} else {
					// Display modified data in JSON message "success"
					c.JSON(http.StatusOK, gin.H{"success": newUser.Desentitize()})
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "role_name field must be admin or user"})
			}
		} else {
			// Display error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "username or password or role_name or rate_formatted field is empty"})
		}
	} else {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
	}

	// curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Thea\", \"lastname\": \"Merlyn\" }" http://localhost:8080/api/v1/users/1
}

func DeleteUser(c *gin.Context) {
	// Get id user
	id := c.Params.ByName("id")
	var user model.User
	// SELECT * FROM users WHERE id = 1;
	model.DB.First(&user, id)

	if user.ID != 0 {
		// DELETE FROM users WHERE id = user.ID
		model.DB.Delete(&user)
		// Display JSON result
		c.JSON(http.StatusOK, gin.H{"success": "User #" + id + " deleted"})
	} else {
		// Display JSON error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
	}

	// curl -i -X DELETE http://localhost:8080/api/v1/users/1
}
