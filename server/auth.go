package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"time"

	"github.com/adamboardman/gracemount/store"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/appleboy/gin-jwt/v3/core"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

type login struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityId = "id"
var identityKey = "email"
var identityConfirmed = "confirmed"

type LoggedInUser struct {
	ID        uint
	Email     string
	Confirmed bool
}

func LogFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func RandomKey(len int) string {
	key := RandomBytes(len)
	return base64.StdEncoding.EncodeToString(key)
}

func RandomBytes(len int) []byte {
	key := make([]byte, len)
	_, err := io.ReadFull(rand.Reader, key)
	LogFatalError(err)
	return key
}

func AllowOptions(c *gin.Context) {
	if c.Request.Method != "OPTIONS" || c.Request.Host != "localhost:3030" {
		c.Next()
	} else {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Content-Type", "application/json")
		c.AbortWithStatus(http.StatusOK)
	}
}

func payloadFunc() func(data any) gojwt.MapClaims {
	return func(data any) gojwt.MapClaims {
		if v, ok := data.(*store.User); ok {
			return gojwt.MapClaims{
				identityId:        v.ID,
				identityKey:       v.Email,
				identityConfirmed: v.Confirmed,
			}
		}
		return gojwt.MapClaims{}
	}
}
func identityHandler() func(c *gin.Context) any {
	return func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return &LoggedInUser{
			ID:        uint(claims[identityId].(float64)),
			Email:     claims[identityKey].(string),
			Confirmed: claims[identityConfirmed] == true,
		}
	}
}

func authenticator(a *WebApp) func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (interface{}, error) {
		var loginVals login
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}

		user, err := a.Store.FindUser(loginVals.Email)
		if err != nil {
			return nil, err
		}
		salt, err := base64.StdEncoding.DecodeString(user.Salt)
		encrypted := argon2.IDKey([]byte(loginVals.Password), salt, 1, 64*1024, 4, 32)
		loginPassword := base64.StdEncoding.EncodeToString(encrypted)
		if err == nil && loginPassword == user.Password {
			return user, nil
		}

		return nil, jwt.ErrFailedAuthentication
	}
}
func authorizer() func(c *gin.Context, data any) bool {
	return func(c *gin.Context, data any) bool {
		user, ok := data.(*LoggedInUser)
		if ok && user.Confirmed {
			return true
		}

		return false
	}
}
func loginResponse() func(c *gin.Context, token *core.Token) {
	return func(c *gin.Context, token *core.Token) {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"token":  token.AccessToken,
			"expire": time.Unix(token.ExpiresAt, 0).Format(time.RFC3339),
		})
	}
}
func (a *WebApp) InitAuth(group *gin.RouterGroup) *jwt.GinJWTMiddleware {
	const secretKeyFileName = "secret_key.txt"
	secretKey, err := os.ReadFile(secretKeyFileName)
	if err != nil {
		secretKey, err = os.ReadFile("../" + secretKeyFileName)
		if err != nil {
			secretKey = []byte(RandomKey(30))
			err = os.WriteFile(secretKeyFileName, secretKey, 0666)
			LogFatalError(err)
		}
	}

	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:           "gracemount",
		Key:             secretKey,
		Timeout:         time.Hour * 24 * 7,
		MaxRefresh:      time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     payloadFunc(),
		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(a),
		Authorizer:      authorizer(),
		LoginResponse:   loginResponse(),
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"status":  code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	// Initialize middleware
	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	auth := group.Group("/auth")
	auth.OPTIONS("/register", AllowOptions)
	auth.OPTIONS("/login", AllowOptions)
	auth.OPTIONS("/refresh_token", AllowOptions)
	auth.POST("/register", RegisterUser)
	auth.POST("/login", authMiddleware.LoginHandler)
	auth.GET("/confirm_email", ConfirmEmail)
	//auth.Use(authMiddleware.MiddlewareFunc())
	//{
	//	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	//}

	return authMiddleware
}

func (a *WebApp) AuthAvailableMiddleware() *jwt.GinJWTMiddleware {
	const secretKeyFileName = "secret_key.txt"
	secretKey, err := os.ReadFile(secretKeyFileName)
	if err != nil {
		secretKey, err = os.ReadFile("../" + secretKeyFileName)
		if err != nil {
			LogFatalError(err)
			return nil
		}
	}

	authAvailMiddleware := &jwt.GinJWTMiddleware{
		Realm:           "gracemount",
		Key:             secretKey,
		Timeout:         time.Hour * 24 * 7,
		MaxRefresh:      time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     payloadFunc(),
		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(a),
		Authorizer:      authorizer(),
		LoginResponse:   loginResponse(),
		DisabledAbort:   true,
		Unauthorized: func(c *gin.Context, code int, message string) {
			// Do nothing as we want some pages to be visible for both all users but still know when we are logged in
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	// Initialize middleware
	errInit := authAvailMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authAvailMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	return authAvailMiddleware
}

type RegisterJSON struct {
	Email                string
	Password             string
	PasswordConfirmation string `json:"password_confirmation"`
	Verification         string
}

func RegisterUser(c *gin.Context) {
	registerJSON := RegisterJSON{}
	if c.ContentType() == "application/json" {
		err := c.BindJSON(&registerJSON)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "JSON Bind failure"})
			return
		}
	} else {
		registerJSON.Email = c.PostForm("email")
		registerJSON.Password = c.PostForm("password")
		registerJSON.PasswordConfirmation = c.PostForm("password_confirmation")
		registerJSON.Verification = c.PostForm("varification")
	}

	if registerJSON.Password != registerJSON.PasswordConfirmation {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Passwords don't match"})
		return
	}

	existingUser, _ := App.Store.FindUser(registerJSON.Email)
	if existingUser != nil {
		salt, _ := base64.StdEncoding.DecodeString(existingUser.Salt)
		verification, _ := base64.StdEncoding.DecodeString(registerJSON.Verification)
		encrypted := argon2.IDKey(verification, salt, 1, 64*1024, 4, 32)
		confirmVerification := base64.StdEncoding.EncodeToString(encrypted)
		if confirmVerification == existingUser.ConfirmVerifier {
			if len(existingUser.Password) == 0 {
				encrypted = argon2.IDKey([]byte(registerJSON.Password), salt, 1, 64*1024, 4, 32)
				existingUser.Password = base64.StdEncoding.EncodeToString(encrypted)
				existingUser.Confirmed = true
				existingUser.ConfirmVerifier = ""

				_, err := App.Store.UpdateUser(existingUser)
				if err == nil {
					c.JSON(http.StatusOK, gin.H{
						"status": http.StatusOK, "message": "User registered successfully", "resourceId": existingUser.ID,
					})
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Verification failure"})
		return
	}

	user := store.User{}
	user.Email = registerJSON.Email
	salt := RandomBytes(16)
	encrypted := argon2.IDKey([]byte(registerJSON.Password), salt, 1, 64*1024, 4, 32)
	user.Salt = base64.StdEncoding.EncodeToString(salt)
	user.Password = base64.StdEncoding.EncodeToString(encrypted)

	verification := RandomBytes(20)
	verificationKey := argon2.IDKey(verification, salt, 1, 64*1024, 4, 32)
	user.ConfirmVerifier = base64.StdEncoding.EncodeToString(verificationKey)

	_, err := App.Store.InsertUser(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "InsertUser failure"})
		return
	}

	SendEmail(registerJSON.Email, base64.StdEncoding.EncodeToString(verification), "", "")

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK, "message": "User registered successfully", "resourceId": user.ID,
	})
}

func InviteUser(email string, invite string, description string) (error, *store.User) {
	user := store.User{}
	user.Email = email
	salt := RandomBytes(16)
	user.Salt = base64.StdEncoding.EncodeToString(salt)

	verification := RandomBytes(20)
	verificationKey := argon2.IDKey(verification, salt, 1, 64*1024, 4, 32)
	user.ConfirmVerifier = base64.StdEncoding.EncodeToString(verificationKey)

	_, err := App.Store.InsertUser(&user)
	if err != nil {
		return err, &user
	}
	SendEmail(email, base64.StdEncoding.EncodeToString(verification), invite, description)

	return nil, &user
}

func SendEmail(emailAddress string, verificationKey string, invite string, description string) {
	data := url.Values{}
	data.Set("email", url.QueryEscape(emailAddress))
	data.Set("verification", url.QueryEscape(verificationKey))
	confirmUrl := "https://gracemount.thinkglobally.org/api/auth/confirm_email?" + data.Encode()
	log.Print(confirmUrl)
	c, err := smtp.Dial("localhost:25")
	if err != nil {
		log.Print(err)
		return
	}
	defer c.Close()
	_ = c.Mail("no-reply@thinkglobally.org")
	_ = c.Rcpt(emailAddress)
	boundary := base64.StdEncoding.EncodeToString(RandomBytes(16))
	wc, err := c.Data()
	LogFatalError(err)
	defer wc.Close()

	subject := "Gracemount Confirm Email Address"
	opening := "Thanks for signing up for a"
	middling := ""
	middlingHTML := ""
	ending := ""
	if len(invite) > 0 {
		subject = "Invite to Gracemount Tree and Plant cataloge"
		opening = "You've been invited open a"
		middling = invite + "\r\nDescription: " + description + "\r\n"
		middlingHTML = "<p>" + invite + "</p>" + "\r\n<p>Description: " + description + "</p>" + "\r\n"
		ending = "and select a password "
	}

	buf := bytes.NewBufferString("" +
		"Subject: " + subject + "\r\n" +
		"From: Gracemount <no-reply@thinkglobally.org>\r\n" +
		"Reply-To: Gracemount <no-reply@thinkglobally.org>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		opening + " Gracemount - Tree and Plant catalogue account\r\n" +
		"\r\n" +
		middling +
		"Please click on the following link to confirm your email address " + ending + confirmUrl + "\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		"<!DOCTYPE html>\r\n" +
		"<html>\r\n" +
		"<head>\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"<p>" + opening + " Gracemount - Tree and Plant catalogue account</p>\r\n" +
		"\r\n" +
		middlingHTML +
		"<p>Please click on the following link to confirm your email address " + ending + "<a href=" + confirmUrl + ">" + confirmUrl + "</a></p>\r\n" +
		"</body>\r\n" +
		"</html>\r\n" +
		"\r\n" +
		"--" + boundary + "--\r\n")
	_, err = buf.WriteTo(wc)
	LogFatalError(err)
}

func ConfirmEmail(c *gin.Context) {
	email, err := url.QueryUnescape(c.Query("email"))
	verificationKey, err := url.QueryUnescape(c.Query("verification"))

	log.Print(email + " " + verificationKey)

	user, err := App.Store.FindUser(email)
	if err == nil {
		salt, _ := base64.StdEncoding.DecodeString(user.Salt)
		verification, _ := base64.StdEncoding.DecodeString(verificationKey)
		encrypted := argon2.IDKey(verification, salt, 1, 64*1024, 4, 32)
		confirmVerification := base64.StdEncoding.EncodeToString(encrypted)
		if confirmVerification == user.ConfirmVerifier {
			if len(user.Password) > 0 {
				user.Confirmed = true
				user.ConfirmVerifier = ""
				_, _ = App.Store.UpdateUser(user)
				c.Redirect(307, "/")
			} else {
				c.Redirect(307, "/register?email="+url.QueryEscape(email)+"&verification="+url.QueryEscape(verificationKey))
			}
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
	}
}
