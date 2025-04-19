// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/edalmava/autoevaluacion/internal/db/models"
	"github.com/golang-jwt/jwt/v5"
)

// Clave secreta para firmar tokens JWT
var jwtSecret = []byte("Edalmava-2025-Autoevaluacion") // Cámbiala en producción

// Claims representa los claims de nuestro JWT
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type contextKey string

const (
	usernameContextKey contextKey = "username"
	roleContextKey     contextKey = "role"
)

// GenerateToken genera un nuevo token JWT para un usuario
func GenerateToken(user models.User) (string, error) {
	// Crear los claims
	claims := &Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token válido por 24 horas
		},
	}

	// Crear el token con el método de firma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token con la clave secreta
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken valida y extrae la información de un token JWT
/* func ParseToken(tokenString string) (*Claims, error) {
	// Parsear el token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
*/
// ParseToken valida y extrae la información de un token JWT
func ParseToken(tokenString string) (*Claims, error) {
	// Crear un objeto Claims vacío para almacenar la información del token
	claims := &Claims{}

	// Parsear el token utilizando ParseWithClaims
	// En v5, el manejo de errores y la estructura es ligeramente diferente
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verificar que el método de firma sea el esperado
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		},
	)

	// Manejar errores de parsing
	if err != nil {
		return nil, err
	}

	// Verificar si el token es válido
	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	// En este punto sabemos que el token es válido y los claims han sido rellenados
	return claims, nil
}

// Auth middleware para validar tokens JWT
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el token del header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// El formato del header debe ser "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}

		// Extraer y validar el token
		tokenString := parts[1]
		claims, err := ParseToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Guardar información del usuario en el contexto
		ctx := context.WithValue(r.Context(), usernameContextKey, claims.Username)
		ctx = context.WithValue(ctx, roleContextKey, claims.Role)

		// Llamar al siguiente handler con el nuevo contexto
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole verifica que el usuario tenga el rol requerido
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(roleContextKey).(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Verificar si el usuario tiene uno de los roles permitidos
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUsernameFromContext extrae el nombre de usuario del contexto
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameContextKey).(string)
	return username, ok
}

// GetRoleFromContext extrae el rol del contexto
func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleContextKey).(string)
	return role, ok
}
