package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"uvcs/handlers"
	"uvcs/modules/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleRegister(t *testing.T) {
	// Setup
	tdb, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", handlers.HandleRegister)

	tests := []struct {
		name        string
		form        url.Values
		wantStatus  int
		wantError   bool
		errorString string
	}{
		{
			name: "valid registration",
			form: url.Values{
				"username":  {"testuser"},
				"email":     {"test@example.com"},
				"password":  {"password123"},
				"firstname": {"Test"},
				"lastname":  {"User"},
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "missing username",
			form: url.Values{
				"email":     {"test@example.com"},
				"password":  {"password123"},
				"firstname": {"Test"},
				"lastname":  {"User"},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "username is required",
		},
		{
			name: "missing email",
			form: url.Values{
				"username":  {"testuser"},
				"password":  {"password123"},
				"firstname": {"Test"},
				"lastname":  {"User"},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "email is required",
		},
		{
			name: "missing password",
			form: url.Values{
				"username":  {"testuser"},
				"email":     {"test@example.com"},
				"firstname": {"Test"},
				"lastname":  {"User"},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "password is required",
		},
		{
			name: "duplicate email",
			form: url.Values{
				"username":  {"testuser2"},
				"email":     {"test@example.com"},
				"password":  {"password123"},
				"firstname": {"Test"},
				"lastname":  {"User"},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "email already exists",
		},
		{
			name: "duplicate username",
			form: url.Values{
				"username":  {"testuser"},
				"email":     {"test2@example.com"},
				"password":  {"password123"},
				"firstname": {"Test"},
				"lastname":  {"User"},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/register", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError {
				assert.Contains(t, w.Body.String(), tt.errorString)
			} else {
				// Verify user was created
				var count int
				err := tdb.DB.QueryRow(`
					SELECT COUNT(*) FROM users 
					WHERE username = $1 AND email = $2
				`, tt.form.Get("username"), tt.form.Get("email")).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	// Setup
	tdb, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Create test user
	_, username, email, _, _ := tdb.CreateTestUser(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", handlers.HandleLogin)

	tests := []struct {
		name        string
		form        url.Values
		wantStatus  int
		wantError   bool
		errorString string
	}{
		{
			name: "valid login with username",
			form: url.Values{
				"identifier": {username},
				"password":   {"password123"},
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "valid login with email",
			form: url.Values{
				"identifier": {email},
				"password":   {"password123"},
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "missing identifier",
			form: url.Values{
				"password": {"password123"},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "identifier is required",
		},
		{
			name: "missing password",
			form: url.Values{
				"identifier": {username},
			},
			wantStatus:  http.StatusBadRequest,
			wantError:   true,
			errorString: "password is required",
		},
		{
			name: "invalid identifier",
			form: url.Values{
				"identifier": {"nonexistent"},
				"password":   {"password123"},
			},
			wantStatus:  http.StatusUnauthorized,
			wantError:   true,
			errorString: "invalid credentials",
		},
		{
			name: "invalid password",
			form: url.Values{
				"identifier": {username},
				"password":   {"wrongpassword"},
			},
			wantStatus:  http.StatusUnauthorized,
			wantError:   true,
			errorString: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/login", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError {
				assert.Contains(t, w.Body.String(), tt.errorString)
			} else {
				// Verify response contains auth keys
				assert.Contains(t, w.Body.String(), "skey1")
				assert.Contains(t, w.Body.String(), "skey2")
			}
		})
	}
}
